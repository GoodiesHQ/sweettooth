package choco

import (
	"os/exec"
	"strings"

	"github.com/goodieshq/sweettooth/pkg/util"
	"github.com/rs/zerolog/log"
)

type PkgAction uint8

const (
	PKG_ACTION_INSTALL   PkgAction = 1
	PKG_ACTION_UPGRADE   PkgAction = 2
	PKG_ACTION_UNINSTALL PkgAction = 3
)

type PackageParams struct {
	Action           PkgAction // (required) install, uninstall, upgrade
	Name             string    // (required) name of the chocolatey package
	Version          string    // (optional) target version to install or upgrade to
	IgnoreChecksum   bool      // (optional) ignore the checksum validation
	InstallOnUpgrade bool      // true = install during upgrade job if not installed, false = skip
	Force            bool      // Force the behavior
	Verbose          bool      // verbose output
	NotSilent        bool      // enable to prevent silent installations
}

type PackageResult struct {
	Params   PackageParams
	Err      error
	Status   ChocoStatus
	Output   string
	ExitCode int
}

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

func ListChocoInstalled() ([]util.Software, error) {
	// choco list -r
	cmd := exec.Command("choco", "list", "-r")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	var packages []util.Software
	lines := strings.Split(string(output), "\n")

	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}

	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		packages = append(packages, util.Software{
			Name:    parts[0],
			Version: parts[1],
		})
	}

	return packages, nil
}

func Package(params *PackageParams) *PackageResult {
	var result = &PackageResult{Params: *params}
	var err error

	var args = []string{pkgactionName(params.Action), params.Name, "-y"}

	// Add on optional arguments based on the InstallParams provided

	if params.Verbose {
		args = append(args, "--verbose")
	}

	if params.Force {
		args = append(args, "--force")
	}

	if (params.Action == PKG_ACTION_INSTALL || params.Action == PKG_ACTION_UPGRADE) && params.Version != "" {
		args = append(args, "--version", params.Version)
	}

	if params.IgnoreChecksum {
		args = append(args, "--ignore-checksums")
	}

	if (params.Action == PKG_ACTION_UPGRADE) && !params.InstallOnUpgrade {
		args = append(args, "--fail-on-not-installed")
	}

	// Run the command
	cmd := exec.Command("choco", args...)
	log.Info().Str("command", strings.Join(cmd.Args, " ")).Send()

	// retrieve the stdout and stderr output
	output, err := cmd.CombinedOutput()
	log.Info().Msgf("Output is %d bytes", len(output))

	result.Status = StatusCheck(output)
	result.Output = string(output)

	if exitError, ok := err.(*exec.ExitError); ok {
		// The program has exited with a non-zero status
		code := exitError.ExitCode()
		result.ExitCode = code
		result.Err = err
	} else if err != nil {
		// There was an error running the command
		result.Err = err
		result.ExitCode = -1
	} else {
		// The program has exited with a zero status
		result.ExitCode = 0
		result.Err = nil
	}

	return result
}
