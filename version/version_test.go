package version

import (
	"fmt"
	"reflect"
	"testing"
)

func TestNewInfo(t *testing.T) {
	v := NewInfo("a", "b", "c")
	if len(v.UUID) == 0 {
		t.Errorf("version uuid should not be empty")
	}
	if v.SemVer != "a" {
		t.Errorf("version semver expected: %s, actual: %s", "a", v.SemVer)
	}
	if v.GitCommit != "b" {
		t.Errorf("version git commit expected: %s, actual: %s", "b", v.GitCommit)
	}
	if v.BuildDate != "c" {
		t.Errorf("version build date expected: %s, actual: %s", "c", v.BuildDate)
	}
}

func TestFormatted(t *testing.T) {
	v := &Info{"a", "b", "c", "d"}
	if v.Formatted() != fmt.Sprintf("&version.Info{SemVer:\"%s\", GitCommit:\"%s\", BuildDate:\"%s\"}", v.SemVer, v.GitCommit, v.BuildDate) {
		t.Errorf("formatted version info incorrect")
	}
}

func TestUpdate(t *testing.T) {
	v1 := NewInfo("a", "b", "c")
	v2 := NewInfo("d", "e", "f")

	v1.Update(v2)
	if !reflect.DeepEqual(v1, v2) {
		t.Errorf("versions should be equal")
	}
}
