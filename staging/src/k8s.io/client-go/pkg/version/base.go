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
	gitVersion   = "v1.31.0-k3s1"
	gitCommit    = "e69f2ced3946a6a4ef56d99c60b799a9817f115a"
	gitTreeState = "clean"
	buildDate = "2024-08-13T21:51:07Z"
)
