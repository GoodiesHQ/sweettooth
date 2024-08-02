package util

import "regexp"

// Simple software application definition
type Software struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Software which has a newer version available
type SoftwareOutdated struct {
	Name       string `json:"name"`
	VersionOld string `json:"versionOld"`
	VersionNew string `json:"versionNew"`
	Pinned     bool   `json:"pinned"`
}

type SoftwareMap struct {
	Regex   regexp.Regexp // a regex to check against the system-reported software name
	Package string        // the corresponding chocolatey package that should be targetted upon a match
}

type Repository struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	Disabled    bool   `json:"disabled"`
	Username    string `json:"username"`
	Certificate string `json:"certificate"`
	Priority    int    `json:"priority"`
	BypassProxy bool   `json:"bypass_proxy"`
	SelfService bool   `json:"self_service"`
	AdminOnly   bool   `json:"admin_only"`
}
