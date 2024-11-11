package choco

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

const (
	TIMEOUT_CHECK_INSTALL = time.Second * 1
	TIMEOUT_CHOCO_INSTALL = time.Minute * 15
)

// create powershell script to install the current version of chocolatey
const installScript string = "" +
	`Set-ExecutionPolicy Bypass -Scope Process -Force;` +
	`[System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072;` +
	`iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))`

func Installed() bool {
	// run choco -v to determine if it is installed or not
	ctx, cancel := context.WithTimeout(context.Background(), TIMEOUT_CHECK_INSTALL)
	defer cancel()
	return exec.CommandContext(ctx, "choco", "-v").Run() == nil
}

func InstallChocolatey() error {
	// call powershell with the installation script
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "powershell", "-Command", installScript)

	// get the output of the installation
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install Chocolatey: %w\nOutput: %s", err, string(output))
	}

	return nil
}

func Bootstrap() error {
	if Installed() {
		// chocolatey is already installed, nothing is needed
		return nil
	}

	// powershell command to install chocolatey
	return InstallChocolatey()
}
