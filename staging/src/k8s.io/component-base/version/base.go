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
	gitVersion   = "v1.31.1-k3s1"
	gitCommit    = "a21493eeac2053bf893a86bda188d46b0067cb01"
	gitTreeState = "clean"
	buildDate = "2024-09-13T05:33:58Z"
)
