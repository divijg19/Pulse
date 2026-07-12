package version

// Version is the semantic version of the Pulse build. It is overridden at
// build time via -ldflags "-X github.com/divijg19/Pulse/internal/version.Version=vX.Y.Z".
var Version = "dev"

// Commit is the git commit the binary was built from. Overridden at build time.
var Commit = "unknown"
