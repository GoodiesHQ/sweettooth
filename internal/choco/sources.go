package choco

import (
	"os/exec"
	"reflect"
	"strconv"
	"strings"

	"github.com/goodieshq/sweettooth/internal/util"
	"github.com/rs/zerolog/log"
)

type SrcAction uint8

const (
	SRC_ACTION_ADD     SrcAction = 1
	SRC_ACTION_REMOVE  SrcAction = 2
	SRC_ACTION_ENABLE  SrcAction = 3
	SRC_ACTION_DISABLE SrcAction = 4
)

type SourceParams struct {
	Action      SrcAction
	Name        string
	URL         string
	Username    string
	Password    string
	Credential  string // TODO: properly utilize this
	Priority    int
	BypassProxy bool
	SelfService bool
	AdminOnly   bool
}

func ListSources() ([]util.Repository, error) {
	// choco sources list -r
	cmd := exec.Command("choco", "sources", "list", "-r")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	// create a list of repositories to add the parsed entries
	var repositories []util.Repository
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// sources list truncated format is well-defined
		parts := strings.Split(line, "|")
		if len(parts) != reflect.TypeFor[util.Repository]().NumField() {
			continue
		}

		/* parse bools and ints from the output */

		disabled, err := strconv.ParseBool(parts[2])
		if err != nil {
			log.Warn().Err(err).Send()
			continue
		}

		priority, err := strconv.Atoi(parts[5])
		if err != nil {
			log.Warn().Err(err).Send()
			continue
		}

		bypassProxy, err := strconv.ParseBool(parts[6])
		if err != nil {
			log.Warn().Err(err).Send()
			continue
		}

		selfService, err := strconv.ParseBool(parts[7])
		if err != nil {
			log.Warn().Err(err).Send()
			continue
		}

		adminOnly, err := strconv.ParseBool(parts[8])
		if err != nil {
			log.Warn().Err(err).Send()
			continue
		}

		repositories = append(repositories, util.Repository{
			Name:        parts[0],
			URL:         parts[1],
			Disabled:    disabled,
			Username:    parts[3],
			Certificate: parts[4],
			Priority:    priority,
			BypassProxy: bypassProxy,
			SelfService: selfService,
			AdminOnly:   adminOnly,
		})
	}

	return repositories, nil
}
