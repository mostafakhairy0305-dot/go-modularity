// Package version exposes the tool version, settable at link time:
//
//	go build -ldflags "-X .../internal/shared/version.Version=v1.2.3"
package version

// Version is the tool version string. Defaults to "dev" for source builds.
var Version = "dev"
