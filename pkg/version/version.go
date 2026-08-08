// Build version information injected at build time via ldflags.
package version

import (
	"runtime/debug"
)

var (
	Version   = "development"
	BuildTime = "unknown"
)

func GetVersion() string {
	return Version
}

func GetRevision() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}
	for _, setting := range info.Settings {
		if setting.Key == "vcs.revision" {
			revision := setting.Value
			for _, s := range info.Settings {
				if s.Key == "vcs.modified" && s.Value == "true" {
					revision += "-dirty"
				}
			}
			return revision
		}
	}
	return "unknown"
}

func GetBuildTime() string {
	return BuildTime
}
