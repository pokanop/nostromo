package version

import (
	"fmt"

	"github.com/google/uuid"
)

// Info identifying version information for releases
type Info struct {
	UUID      string `json:"uuid"`
	SemVer    string `json:"semVer"`
	GitCommit string `json:"gitCommit"`
	BuildDate string `json:"buildDate"`
}

// Create a new version info with unique identifier
func NewInfo(semver string, gitCommit string, buildDate string) *Info {
	return &Info{uuid.NewString(), semver, gitCommit, buildDate}
}

// Formatted returns version formatted string
func (i *Info) Formatted() string {
	return fmt.Sprintf("&version.Info{SemVer:\"%s\", GitCommit:\"%s\", BuildDate:\"%s\"}", i.SemVer, i.GitCommit, i.BuildDate)
}

// Update the version info including unique identifier
func (i *Info) Update(ver *Info) {
	i.UUID = ver.UUID
	i.SemVer = ver.SemVer
	i.GitCommit = ver.GitCommit
	i.BuildDate = ver.BuildDate
}
