package system

import (
	"fmt"
	"os"
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

type SystemInfo struct {
	Hostname string
	OSInfo   OSInfo
}

func GetSystemInfo() (*SystemInfo, error) {
	log.Trace().Msg("GetSystemInfo called")
	var sysinfo SystemInfo
	var osinfo *OSInfo = &sysinfo.OSInfo

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	sysinfo.Hostname = hostname

	log.Debug().Msg("Opening CurrentVersion registry")
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE)
	if err != nil {
		return nil, err
	}
	defer key.Close()

	currentVersion, _, err := key.GetStringValue("CurrentVersion")
	if err != nil {
		return nil, err
	}
	osinfo.Kernel = currentVersion

	productName, _, err := key.GetStringValue("ProductName")
	if err != nil {
		return nil, err
	}
	osinfo.Name = productName

	major, _, err := key.GetIntegerValue("CurrentMajorVersionNumber")
	if err != nil {
		return nil, err
	}
	osinfo.Major = int(major)

	minor, _, err := key.GetIntegerValue("CurrentMinorVersionNumber")
	if err != nil {
		return nil, err
	}
	osinfo.Minor = int(minor)

	build, _, err := key.GetStringValue("CurrentBuild")
	if err != nil {
		return nil, err
	}
	buildNumber, err := strconv.Atoi(build)
	if err != nil {
		buildNumber = 0
	}
	osinfo.Build = buildNumber

	// check for windows 11 build numbers
	if major == 10 && buildNumber >= 22000 {
		osinfo.Name = strings.Replace(osinfo.Name, "indows 10", "indows 11", -1)
	}

	osinfo.IsServer = strings.Contains(strings.ToLower(osinfo.Name), "server")
	log.Debug().Str("hostname", sysinfo.Hostname).Str("osname", osinfo.Name).Str("version", fmt.Sprintf("%v.%v.%v", osinfo.Major, osinfo.Minor, osinfo.Build)).Msg("Got System Info")

	return &sysinfo, nil
}
