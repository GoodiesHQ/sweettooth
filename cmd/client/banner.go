package main

import (
	"fmt"

	"github.com/goodieshq/sweettooth/pkg/config"
	"github.com/rs/zerolog/log"
)

const (
	bannerText = "" +
		`  ___                _  _____         _   _    ` + "\n" +
		` / __|_ __ _____ ___| ||_   _|__  ___| |_| |_  ` + "\n" +
		` \__ \ V  V / -_) -_)  _|| |/ _ \/ _ \  _| ' \ ` + "\n" +
		` |___/\_/\_/\___\___|\__||_|\___/\___/\__|_||_|` + "\n" +
		"\n"
)

func banner() {
	fmt.Print(bannerText)
	log.Info().Msg("Initializing " + config.APP_NAME + "...")
	fmt.Println()
}
