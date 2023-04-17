package driver

var (
	DefaultMode              = "unspecified"
	DefaultCSISocket         = "unix:///var/lib/csi/sockets/pluginproxy/csi.sock"
	DefaultNodeID            = ""
	DefaultRootDir           = "/"
	DefaultPluginDataDir     = "/"
	DefaultHandlersCount     = 1
	DefaultSanityTestRun     = false
	DefaultLogLevel          = 5
	DefaultEnforceNamespaces = true
)
