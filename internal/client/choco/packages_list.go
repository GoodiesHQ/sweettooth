package choco

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/goodieshq/sweettooth/internal/util"
	"github.com/rs/zerolog/log"
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

// extract the list of software from a `choco list --include-programs` part?
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

func ListAllInstalled(ctx context.Context) (util.SoftwareList, util.SoftwareList, error) {
	log := util.Logger("choco::ListAllInstalled")
	log.Trace().Msg("called")
	defer log.Trace().Msg("finish")

	log.Debug().Str("cmd", "choco list --include-programs").Msg("listing all installed programs")

	cmd := exec.CommandContext(ctx, "choco", "list", "--include-programs")

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

func ListChocoOutdated(ctx context.Context) (util.SoftwareOutdatedList, error) {
	log := util.Logger("choco::ListChocoOutdated")
	log.Trace().Msg("called")
	defer log.Trace().Msg("finish")

	log.Debug().Str("cmd", "choco outdated -r").Msg("listing outdated choco packages")

	cmd := exec.CommandContext(ctx, "choco", "outdated", "-r")

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

func ListChocoInstalled(ctx context.Context) ([]util.Software, error) {
	log := util.Logger("choco::ListChocoInstalled")
	log.Trace().Msg("called")
	defer log.Trace().Msg("finish")

	log.Debug().Str("cmd", "choco list -r").Msg("listing installed choco packages")

	cmd := exec.CommandContext(ctx, "choco", "list", "-r")

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
