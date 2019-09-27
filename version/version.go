package version

import "fmt"

// Info identifying version information for releases
type Info struct {
	SemVer    string
	GitCommit string
	BuildDate string
}

// Formatted returns version formatted string
func (i *Info) Formatted() string {
	return fmt.Sprintf("&version.Info{SemVer:\"%s\", GitCommit:\"%s\", BuildDate:\"%s\"}", i.SemVer, i.GitCommit, i.BuildDate)
}
