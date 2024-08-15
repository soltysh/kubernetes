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
	gitVersion   = "v1.31.0-k3s2"
	gitCommit    = "ce17ea941c9652287716c6dc5518fd2ef95b2052"
	gitTreeState = "clean"
	buildDate = "2024-08-15T20:35:15Z"
)
