// Package version provides build-time version information.
package version

import "runtime/debug"

const devVersion = "dev"

// Set via ldflags at build time. Takes precedence over gitVersion.
var ldflagsVersion string

// String returns the version string.
// Priority: ldflags > gitVersion (from go:generate) > debug.ReadBuildInfo.
func String() string {
	if ldflagsVersion != "" {
		return ldflagsVersion
	}

	if gitVersion != "" && gitVersion != devVersion {
		return gitVersion
	}

	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		if buildInfo.Main.Version != "" && buildInfo.Main.Version != "(devel)" {
			return buildInfo.Main.Version
		}

		for _, s := range buildInfo.Settings {
			if s.Key == "vcs.tag" && s.Value != "" {
				return s.Value
			}
		}
	}

	return devVersion
}
