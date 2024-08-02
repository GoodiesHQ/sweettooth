package system

import (
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
	"golang.org/x/sys/windows/registry"
)

type OSInfo struct {
	Kernel   string
	Name     string
	IsServer bool
	Major    int
	Minor    int
	Build    int
}

func GetSystemInfo() (*OSInfo, error) {
	var info OSInfo

	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE)
	if err != nil {
		return nil, err
	}
	defer key.Close()

	currentVersion, _, err := key.GetStringValue("CurrentVersion")
	if err != nil {
		return nil, err
	}
	info.Kernel = currentVersion

	productName, _, err := key.GetStringValue("ProductName")
	if err != nil {
		return nil, err
	}
	info.Name = productName

	major, _, err := key.GetIntegerValue("CurrentMajorVersionNumber")
	if err != nil {
		return nil, err
	}
	info.Major = int(major)

	minor, _, err := key.GetIntegerValue("CurrentMinorVersionNumber")
	if err != nil {
		return nil, err
	}
	info.Minor = int(minor)

	build, _, err := key.GetStringValue("CurrentBuild")
	if err != nil {
		return nil, err
	}
	buildNumber, err := strconv.Atoi(build)
	if err != nil {
		buildNumber = 0
	}
	info.Build = buildNumber

	if major == 10 && buildNumber >= 22000 {
		info.Name = strings.Replace(info.Name, "indows 10", "indows 11", -1)
	}

	info.IsServer = strings.Contains(strings.ToLower(info.Name), "server")
	log.Info().Str("osname", info.Name).Msg("Got System Info")

	return &info, nil
}
