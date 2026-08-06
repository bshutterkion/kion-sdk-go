package kion

// Version is this SDK's own release version, shared by every language lane
// (Go, Python, TypeScript) — they are generated from one spec and released
// together. Stamped from the repo-root VERSION file by scripts/set-release.sh;
// `make check-release` fails if it drifts.
//
// Note this is the SDK's version, NOT a Kion API version. Go module consumers
// resolve versions from git tags on the published mirror (go.mod has no version
// field); this constant just lets a program introspect what it was built with.
const Version = "0.8.0"

// KionVersions maps each generated sub-package to the Kion release its API
// belongs to. The SDK ships every supported Kion version simultaneously, which
// is why the SDK's own Version above cannot be a Kion version — one artifact
// spans them all, and a single consumer routinely imports several at once.
var KionVersions = map[string]string{
	"v3_12":  "3.12",
	"v3_13":  "3.13",
	"v3_14":  "3.14",
	"v3_15":  "3.15",
	"v3_16":  "3.16",
	"master": "unreleased (portal/master — internal/dev only, never shipped to customers)",
}

// KionVersion returns the Kion release for a generated sub-package name (e.g.
// "v3_16" -> "3.16"), and whether that sub-package is known.
func KionVersion(subPackage string) (string, bool) {
	v, ok := KionVersions[subPackage]
	return v, ok
}
