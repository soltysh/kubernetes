package version

const (
	// DefaultKubeBinaryVersion is the hard coded k8 binary version based on the latest K8s release.
	// It is supposed to be consistent with gitMajor and gitMinor, except for local tests, where gitMajor and gitMinor are "".
	// Should update for each minor release!
	DefaultKubeBinaryVersion = "1.31"
)

var (
	gitMajor = "1"
	gitMinor = "31"
	gitVersion   = "v1.31.1-k3s2"
	gitCommit    = "e1fe8bca723153dfc8e9959ca3c51625c4fd0a4f"
	gitTreeState = "clean"
	buildDate = "2024-09-13T21:31:30Z"
)
