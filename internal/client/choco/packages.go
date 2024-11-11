package choco

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/goodieshq/sweettooth/pkg/api"
	"github.com/rs/zerolog/log"
)

type PkgAction uint8

const (
	PKG_ACTION_INSTALL   PkgAction = 1
	PKG_ACTION_UPGRADE   PkgAction = 2
	PKG_ACTION_UNINSTALL PkgAction = 3

	PKG_DEFAULT_TIMEOUT = 600
)

func pkgactionName(action PkgAction) string {
	switch action {
	case PKG_ACTION_INSTALL:
		return "install"
	case PKG_ACTION_UPGRADE:
		return "upgrade"
	case PKG_ACTION_UNINSTALL:
		return "uninstall"
	default:
		return ""
	}
}

type partType int

const (
	TYPE_UNKNOWN partType = iota
	TYPE_CHOCO   partType = iota
	TYPE_SYSTEM  partType = iota
)

func streamOutput(pipe io.ReadCloser, buf *bytes.Buffer, name string) {
	buffer := make([]byte, 1024)
	for {
		n, err := pipe.Read(buffer)
		if err != nil {
			if err != io.EOF {
				log.Error().Err(err).Str("name", name).Msg("failed reading pipe")
			}
			break
		}
		if buf != nil && n > 0 {
			fmt.Printf("%s: %s", name, buffer[:n])
			buf.Write(buffer[:n])
		}
	}
}

func addCommandArg(args *[]string, condition bool, value string, additional ...string) {
	if condition {
		if args == nil {
			args = &[]string{}
		}
		*args = append(*args, value)
		if len(additional) > 0 {
			*args = append(*args, additional...)
		}
	}
}

func command(ctx context.Context, cmd string, args ...string) {

}

func Package(ctx context.Context, action PkgAction, params *api.PackageJobParameters) *api.PackageJobResult {
	log.Trace().Msg("choco.Package called")

	var result = &api.PackageJobResult{}
	var err error

	var args = []string{pkgactionName(action), params.Name, "-y", "--no-progress"}

	// Add on optional arguments based on the InstallParams provided

	addCommandArg(&args, params.VerboseOutput, "--verbose") // enable verbose output
	addCommandArg(&args, params.Force, "--force")           // force the command, use sparingly

	addCommandArg(&args, params.IgnoreChecksum, "--ignore-checksums") // ignore checksum, use sparingly

	if slices.Contains([]PkgAction{PKG_ACTION_INSTALL, PKG_ACTION_UPGRADE}, action) {
		addCommandArg(&args, !params.InstallOnUpgrade, "--fail-on-not-installed") // by default, upgrade does not install if missing
		if params.Version != nil && *params.Version != "" {
			addCommandArg(&args, true, "--version", *params.Version) // specify the version of a package on install/upgrade
		}
	}

	if params.Timeout == 0 {
		params.Timeout = PKG_DEFAULT_TIMEOUT
	}

	addCommandArg(&args, true, "--timeout", strconv.Itoa(params.Timeout))

	log.Debug().Str("cmd", "choco "+strings.Join(args, " ")).Msg("running command")

	// Run the command
	ctx, cancel := context.WithTimeout(ctx, time.Duration(params.Timeout)*time.Second+15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "choco", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		result.ExitCode = -1
		msg := fmt.Errorf("failed to get stdout pipe: %w", err).Error()
		result.Error = &msg
		return result
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		result.ExitCode = -1
		msg := fmt.Errorf("failed to get stdout pipe: %w", err).Error()
		result.Error = &msg
		return result
	}

	var bufout, buferr bytes.Buffer

	go streamOutput(stdout, &bufout, "stdout")
	go streamOutput(stderr, &buferr, "stderr")

	// retrieve the stdout and stderr output
	err = cmd.Start()
	if err != nil {
		result.ExitCode = -1
		msg := fmt.Errorf("failed to start the command '%s': %w", "choco "+strings.Join(args, " "), err).Error()
		result.Error = &msg
		return result
	}

	err = cmd.Wait()
	if err != nil {
		if err == context.DeadlineExceeded {
			err = errors.New("the choco job timed out during execution")
		}
		log.Error().Err(err).Msg("failed to get choco output")
	}

	output := append(bufout.Bytes(), buferr.Bytes()...)
	log.Info().Msgf("Output is %d bytes", len(output))

	// set the result
	result.Output = string(output)
	result.Status = int(StatusCheck(output))

	if exitError, ok := err.(*exec.ExitError); ok {
		// The program has exited with a non-zero status
		result.ExitCode = exitError.ExitCode()
		err := err.Error()
		result.Error = &err
	} else if err != nil {
		// There was an error running the command
		result.ExitCode = -1
		err := err.Error()
		result.Error = &err
	} else {
		// The program has exited with a zero status
		result.ExitCode = 0
		result.Error = nil
	}

	return result
}
