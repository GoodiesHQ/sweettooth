package choco

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
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
		if n > 0 {
			fmt.Printf("%s: %s", name, buffer[:n])
			buf.Write(buffer[:n])
		}
	}
}

func Package(ctx context.Context, action PkgAction, params *api.PackageJobParameters) *api.PackageJobResult {
	log.Trace().Msg("choco.Package called")

	var result = &api.PackageJobResult{}
	var err error

	var args = []string{pkgactionName(action), params.Name, "-y", "--no-progress"}

	// Add on optional arguments based on the InstallParams provided

	// enable verbose output
	if params.VerboseOutput {
		args = append(args, "--verbose")
	}

	// force the command, use sparingly
	if params.Force {
		args = append(args, "--force")
	}

	// specify the version of a package on install/upgrade
	if (action == PKG_ACTION_INSTALL || action == PKG_ACTION_UPGRADE) && params.Version != nil && *params.Version != "" {
		args = append(args, "--version", *params.Version)
	}

	// ignore checksum, use sparingly
	if params.IgnoreChecksum {
		args = append(args, "--ignore-checksums")
	}

	// by default, upgrade does not install if missing
	if (action == PKG_ACTION_UPGRADE) && !params.InstallOnUpgrade {
		args = append(args, "--fail-on-not-installed")
	}

	if params.Timeout == 0 {
		params.Timeout = PKG_DEFAULT_TIMEOUT
	}
	args = append(args, fmt.Sprintf("--timeout %d", params.Timeout))

	// Run the command
	log.Debug().Str("cmd", "choco "+strings.Join(args, " ")).Msg("running command")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*(time.Duration(params.Timeout+30))) // add 30s to the context
	defer cancel()

	cmd := exec.CommandContext(ctx, "choco", args...)
	// cmd := exec.Command("choco", args...)

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

	// retrieve the stdout and stderr output
	err = cmd.Start()
	if err != nil {
		result.ExitCode = -1
		msg := fmt.Errorf("failed to start the command '%s': %w", "choco "+strings.Join(args, " "), err).Error()
		result.Error = &msg
		return result
	}

	go streamOutput(stdout, &bufout, "stdout")
	go streamOutput(stderr, &buferr, "stderr")

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
