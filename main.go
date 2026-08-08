// Yunxiao Woodpecker forge addon entry point.
// Reads configuration from environment variables and starts the gRPC addon server.
package main

import (
	"log/slog"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge/addon"
	"k8s.io/utils/env"

	"yunxiao-woodpecker-addon/internal"
	"yunxiao-woodpecker-addon/pkg/version"
)

func main() {
	logLevel := env.GetString("LOG_LEVEL", "info")
	var slogLevel slog.Level
	switch logLevel {
	case "debug":
		slogLevel = slog.LevelDebug
	case "info":
		slogLevel = slog.LevelInfo
	case "warn":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}
	slog.SetLogLoggerLevel(slogLevel)
	slog.Info("yunxiao-woodpecker-addon is starting",
		"version", version.GetVersion(),
		"revision", version.GetRevision(),
		"build_time", version.GetBuildTime())

	opts := internal.ForgeOpts{
		APIURL:         env.GetString("YUNXIAO_API_URL", ""),
		OrganizationID: env.GetString("YUNXIAO_ORGANIZATION_ID", ""),
		WoodpeckerHost: env.GetString("WOODPECKER_HOST", ""),
		HookSecret:     env.GetString("YUNXIAO_HOOK_SECRET", ""),
	}

	f, err := internal.New(opts)
	if err != nil {
		slog.Error("failed to create yunxiao forge", "error", err)
		return
	}
	addon.Serve(f)
}
