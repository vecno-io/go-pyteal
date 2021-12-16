package cfg

type NodeConfig struct {
	Version            uint64
	GossipFanout       uint64
	NetAddress         string
	DNSBootstrapID     string
	EnableProfiler     bool
	EnableDeveloperAPI bool
}
