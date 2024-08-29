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
	gitVersion   = "v1.31.0-k3s3"
	gitCommit    = "e5a2caa3c641c6fb621669b65dc95741c6913ddc"
	gitTreeState = "clean"
	buildDate = "2024-08-29T17:07:24Z"
)
