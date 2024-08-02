package choco

import (
	"regexp"

	"github.com/rs/zerolog/log"
)

// The status is a SweetTooth-specific status code to indicate the result of a Choco job
type ChocoStatus int

/*
 * Statuses:
 * - 10's = install
 * - 20's = upgrade
 * - 30's = uninstall
 *
 * - X[0]   = success as intended
 * - X[1-3] = successes (alternative variations of success)
 * - X[4-6] = neutral (failed to do the task, but it's ok e.g. already installed)
 * - X[7-9] = failures (could not succeed in the task at all)
 *
 */

const (
	StatusUnknownFailure   = -1 // unhandled, but definitely failed
	StatusUnknown          = 0  // default pending
	StatusInstallSuccess   = 10 // successfully installed the target package
	StatusInstallAlready   = 11 // failed to install a package because it already exists
	StatusInstallNoExist   = 17 // failed to install a package because it doesn't exist
	StatusInstallFailure   = 19 // failed to install
	StatusUpgradeSuccess   = 20 // successfully upgraded the target package
	StatusUpgradeAlready   = 21 // failed to upgrade a package because it is already upgraded
	StatusUpgradeNewer     = 22 // failed to upgrade a package because a newer version is installed
	StatusUpgradeNoExist   = 24 // failed to upgrade a package because it isn't installed
	StatusUninstallSuccess = 30 // successfully uninstalled the target package
	StatusUninstallNoExist = 34 // failed to uninstall a package because it isn't installed
	StatusErrorChecksum    = 57 // checksum failed
)

var (
	regexInstallSuccess   = regexp.MustCompile(`(?m)^\s*The install of .* was successful.\s*$`)
	regexInstallAlready   = regexp.MustCompile(`(?m)^\s*- .* - .* v[\d\.]+ already installed.\s*$`)
	regexInstallNoExist   = regexp.MustCompile(`(?m)^\s*- .* - .* not installed\. The package was not found with the source\(s\) listed\.\s*$`)
	regexInstallFailure   = regexp.MustCompile(`(?m)^\s*- .* \(exited 1\) - .* not installed. An error occurred during installation:\s*$`)
	regexUpgradeSuccess   = regexp.MustCompile(`(?m)^\s*The upgrade of .* was successful\.\s*$`)
	regexUpgradeAlready   = regexp.MustCompile(`(?m)^\s*.* v[\d\.]+ is the latest version available based on your source\(s\)\.\s*$`)
	regexUpgradeNewer     = regexp.MustCompile(`(?m)^\s*- .* - A newer version of .* \(v[\d\.]+\) is already installed\.\s*$`)
	regexUpgradeNoExist   = regexp.MustCompile(`(?m)^\s*- .* - .* is not installed. Cannot upgrade a non-existent package\.\s*$`)
	regexUninstallSuccess = regexp.MustCompile(`(?m)^\s*.* has been successfully uninstalled\.\s*$`)
	regexUninstallNoExist = regexp.MustCompile(`(?m)^\s*- .* - .* is not installed\. Cannot uninstall a non-existent package\.\s*$`)
	/* Errors */
	regexErrorChecksum = regexp.MustCompile(`(?m)^ERROR: Checksum for '.*' did not meet '[0-9a-f]+' for checksum type`)
)

var statusPatterns = map[*regexp.Regexp]ChocoStatus{
	regexInstallSuccess:   StatusInstallSuccess,
	regexInstallAlready:   StatusInstallAlready,
	regexInstallNoExist:   StatusInstallNoExist,
	regexInstallFailure:   StatusInstallFailure,
	regexUninstallSuccess: StatusUninstallSuccess,
	regexUninstallNoExist: StatusUninstallNoExist,
	regexUpgradeSuccess:   StatusUpgradeSuccess,
	regexUpgradeAlready:   StatusUpgradeAlready,
	regexUpgradeNoExist:   StatusUpgradeNoExist,
	regexUpgradeNewer:     StatusUpgradeNewer,
	regexErrorChecksum:    StatusErrorChecksum,
}

var statusMessages = map[ChocoStatus]string{
	StatusInstallSuccess:   "successfully installed",
	StatusInstallAlready:   "already installed",
	StatusInstallNoExist:   "unknown package",
	StatusInstallFailure:   "installation failure",
	StatusUninstallSuccess: "successfully uninstalled",
	StatusUninstallNoExist: "package not installed",
	StatusUpgradeSuccess:   "successfully upgraded",
	StatusUpgradeAlready:   "already upgraded",
	StatusUpgradeNoExist:   "package not installed",
	StatusUpgradeNewer:     "newer version installed",
	StatusErrorChecksum:    "invalid checksum",
	StatusUnknownFailure:   "unspecified failure",
}

// return a human-friendly message from a status code
func StatusMessage(s ChocoStatus) string {
	if msg, found := statusMessages[s]; !found {
		return "unknown status"
	} else {
		return msg
	}
}

// apply various regex patterns to the output of Chocolatey to determine the status of the command
func StatusCheck(output []byte) (status ChocoStatus) {
	status = StatusUnknownFailure
	for pattern, s := range statusPatterns {
		if pattern.Find(output) != nil {
			status = s
			break
		}
	}

	log.Debug().Int("status", int(status)).Str("message", StatusMessage(status)).Send()
	return
}
