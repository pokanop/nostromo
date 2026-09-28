package model

import (
	"fmt"
	"os"
	"runtime"
	"sort"
	"strings"
)

// PlatformEnv overrides the detected platform, e.g. NOSTROMO_PLATFORM=windows
// or NOSTROMO_PLATFORM=linux/arm64. Useful for testing manifests targeting
// other platforms.
const PlatformEnv = "NOSTROMO_PLATFORM"

// Known GOOS and GOARCH values, see `go tool dist list`
var (
	knownOS = map[string]bool{
		"aix": true, "android": true, "darwin": true, "dragonfly": true, "freebsd": true,
		"illumos": true, "ios": true, "js": true, "linux": true, "netbsd": true,
		"openbsd": true, "plan9": true, "solaris": true, "wasip1": true, "windows": true,
	}
	knownArch = map[string]bool{
		"386": true, "amd64": true, "arm": true, "arm64": true, "loong64": true,
		"mips": true, "mips64": true, "mips64le": true, "mipsle": true, "ppc64": true,
		"ppc64le": true, "riscv64": true, "s390x": true, "wasm": true,
	}
)

// CurrentPlatform returns the platform commands are filtered against as
// "os/arch", honoring the NOSTROMO_PLATFORM override which may omit the arch.
func CurrentPlatform() string {
	if p := strings.ToLower(strings.TrimSpace(os.Getenv(PlatformEnv))); len(p) > 0 {
		return p
	}
	return runtime.GOOS + "/" + runtime.GOARCH
}

// ParsePlatforms splits a comma separated platform list, validating and
// normalizing each entry. An empty list means all platforms.
func ParsePlatforms(list []string) ([]string, error) {
	platforms := []string{}
	seen := map[string]bool{}
	for _, item := range list {
		for _, p := range strings.Split(item, ",") {
			p = strings.ToLower(strings.TrimSpace(p))
			if len(p) == 0 {
				continue
			}
			if err := ValidatePlatform(p); err != nil {
				return nil, err
			}
			if !seen[p] {
				seen[p] = true
				platforms = append(platforms, p)
			}
		}
	}
	return platforms, nil
}

// ValidatePlatform checks that a platform is a known GOOS or GOOS/GOARCH pair
func ValidatePlatform(platform string) error {
	goos, goarch, hasArch := strings.Cut(platform, "/")
	if !knownOS[goos] {
		return fmt.Errorf("invalid platform %q, os must be one of [%s]", platform, strings.Join(sortedKeys(knownOS), ", "))
	}
	if hasArch && !knownArch[goarch] {
		return fmt.Errorf("invalid platform %q, arch must be one of [%s]", platform, strings.Join(sortedKeys(knownArch), ", "))
	}
	return nil
}

// SupportedOS returns the known GOOS values
func SupportedOS() []string {
	return sortedKeys(knownOS)
}

// platformMatches reports whether a platform spec ("os" or "os/arch") covers
// the given platform. An arch is only compared when both sides specify one.
func platformMatches(spec, platform string) bool {
	specOS, specArch, specHasArch := strings.Cut(spec, "/")
	goos, goarch, hasArch := strings.Cut(platform, "/")
	if specOS != goos {
		return false
	}
	if specHasArch && hasArch && specArch != goarch {
		return false
	}
	return true
}

// platformsInclude reports whether a platform list covers the given
// platform. An empty list covers every platform.
func platformsInclude(platforms []string, platform string) bool {
	if len(platforms) == 0 {
		return true
	}
	for _, spec := range platforms {
		if platformMatches(spec, platform) {
			return true
		}
	}
	return false
}

func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
