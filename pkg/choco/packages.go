package choco

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/goodieshq/sweettooth/pkg/api"
	"github.com/goodieshq/sweettooth/pkg/util"
	"github.com/rs/zerolog/log"
)

type PkgAction uint8

const (
	PKG_ACTION_INSTALL   PkgAction = 1
	PKG_ACTION_UPGRADE   PkgAction = 2
	PKG_ACTION_UNINSTALL PkgAction = 3
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

// convert an arbitrarily delimited lines into util.Software name/version
func parseSoftwareList(lines []string, delim string) util.SoftwareList {
	var software util.SoftwareList
	for _, line := range lines {
		parts := strings.Split(line, delim)
		if len(parts) != 2 { // line is unexpected
			log.Warn().Str("line", line).Msg("invalid line")
			continue
		}
		software = append(software, util.Software{Name: parts[0], Version: parts[1]})
	}
	return software
}

func getListAllInstalledPart(part string) (util.SoftwareList, partType) {
	lines := strings.Split(strings.TrimSpace(part), "\n")
	if len(lines) > 0 {
		if strings.HasSuffix(strings.ToLower(lines[len(lines)-1]), "packages installed.") {
			return parseSoftwareList(lines[0:len(lines)-1], " "), TYPE_CHOCO
		}
		if strings.HasSuffix(strings.ToLower(lines[len(lines)-1]), "applications not managed with chocolatey.") {
			return parseSoftwareList(lines[0:len(lines)-1], "|"), TYPE_SYSTEM
		}
	}
	return nil, TYPE_UNKNOWN
}

func ListAllInstalled() (util.SoftwareList, util.SoftwareList, error) {
	log.Trace().Msg("choco.ListChocoOutdated called")
	log.Debug().Str("cmd", "choco list --include-programs").Msg("running command")

	cmd := exec.Command("choco", "list", "--include-programs")

	output, err := cmd.Output()
	if err != nil {
		return nil, nil, err
	}

	// strip the first line, it only causes problems
	if bytes.HasPrefix(output, []byte("Chocolatey ")) {
		if i := bytes.IndexByte(output, '\n'); i >= 0 {
			output = output[i+1:]
		}
	}

	// split the output into sections separated by two newlines
	parts := strings.Split(strings.ReplaceAll(string(output), "\r\n", "\n"), "\n\n")

	if len(parts) < 2 {
		return nil, nil, fmt.Errorf("unexpected output '%s'", string(output))
	}

	var pkgChoco, pkgSystem util.SoftwareList = nil, nil

	for _, part := range parts {
		software, partType := getListAllInstalledPart(part)
		switch partType {
		case TYPE_CHOCO:
			pkgChoco = software
		case TYPE_SYSTEM:
			pkgSystem = software
		default:
			continue
		}
	}

	if pkgChoco == nil {
		return nil, nil, fmt.Errorf("unable to find choco output")
	}

	if pkgSystem == nil {
		return nil, nil, fmt.Errorf("unable to find system output")
	}

	return pkgChoco, pkgSystem, nil
}

func ListChocoOutdated() (util.SoftwareOutdatedList, error) {
	log.Trace().Msg("choco.ListChocoOutdated called")
	log.Debug().Str("cmd", "choco outdated -r").Msg("running command")

	cmd := exec.Command("choco", "outdated", "-r")

	// get the stdout output
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var packages = []util.SoftwareOutdated{}

	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		// trim any remaining whitespace
		line = strings.TrimSpace(line)

		// ignore blank lines
		if line == "" {
			continue
		}

		// parse the pipe-delimited output
		parts := strings.Split(line, "|")
		if len(parts) != 4 {
			log.Warn().Str("line", line).Msg("failed to split line by '|'")
			continue
		}

		pinned, err := strconv.ParseBool(parts[3])
		if err != nil {
			log.Warn().Str("line", line).Msg("boolean value is invalid")
			continue
		}

		// add the package described in the line to the list
		packages = append(packages, util.SoftwareOutdated{
			Name:       parts[0],
			VersionOld: parts[1],
			VersionNew: parts[2],
			Pinned:     pinned,
		})
	}

	return packages, nil
}

func ListChocoInstalled() ([]util.Software, error) {
	log.Trace().Msg("choco.ListChocoInstalled called")
	log.Debug().Str("cmd", "choco list -r").Msg("running command")

	cmd := exec.Command("choco", "list", "-r")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	var packages = []util.Software{}
	lines := strings.Split(string(output), "\n")

	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}

	for _, line := range lines {
		if line == "" {
			continue
		}
		// parse the pipe-delimited output
		parts := strings.Split(line, "|")
		if len(parts) != 2 {
			log.Warn().Str("line", line).Msg("failed to split line by '|'")
			continue
		}

		packages = append(packages, util.Software{
			Name:    parts[0],
			Version: parts[1],
		})
	}

	return packages, nil
}

func Package(action PkgAction, params *api.PackageJobParameters) *api.PackageJobResult {
	log.Trace().Msg("choco.Package called")

	var result = &api.PackageJobResult{}
	var err error

	var args = []string{pkgactionName(action), params.Name, "-y"}

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

	// Run the command
	log.Debug().Str("cmd", "choco "+strings.Join(args, " ")).Msg("running command")
	cmd := exec.Command("choco", args...)

	// retrieve the stdout and stderr output
	output, err := cmd.CombinedOutput()
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
