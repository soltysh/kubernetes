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
	gitVersion   = "v1.31.1-k3s3"
	gitCommit    = "94d3e600eedb04b7612a9bda3bf1d348e5dd769e"
	gitTreeState = "clean"
	buildDate = "2024-09-17T18:17:10Z"
)
