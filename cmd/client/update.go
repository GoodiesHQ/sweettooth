package main

import (
	"context"
	"errors"
	"fmt"
	"runtime"

	"github.com/creativeprojects/go-selfupdate"
	"github.com/goodieshq/sweettooth/pkg/info"
	"github.com/rs/zerolog/log"
)

func update(version string) (bool, error) {
	latest, found, err := selfupdate.DetectLatest(context.Background(), selfupdate.ParseSlug("goodieshq/sweettooth"))
	if err != nil {
		return false, fmt.Errorf("error occurred while detecting version: %w", err)
	}
	if !found {
		return false, fmt.Errorf("latest version for %s/%s could not be found from github repository", runtime.GOOS, runtime.GOARCH)
	}

	log.Info().Str("current", version).Str("latest", latest.Version()).Msg("detected versions")
	if latest.LessOrEqual(version) {
		log.Info().Msg(info.APP_NAME + " client is already up to date")
		return false, nil
	}

	exe, err := selfupdate.ExecutablePath()
	if err != nil {
		return false, errors.New("could not locate executable path")
	}
	if err := selfupdate.UpdateTo(context.Background(), latest.AssetURL, latest.AssetName, exe); err != nil {
		return false, fmt.Errorf("error occurred while updating binary: %w", err)
	}

	log.Info().Msgf("successfully updated from %s -> %s", version, latest.Version())
	return true, nil
}
