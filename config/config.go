package config

// Config represents configuration for the memap core.
type Config struct {
	CleanerInterval int             `yaml:"cleanerInterval"`
	Namespace       NamespaceConfig `yaml:"namespace"`
}

// NamespaceConfig represents configuration for namespaces.
type NamespaceConfig struct {
	ShardCounts ShardCounts `yaml:"shardCounts"`
}

// ShardCounts specifies the number of shards for each storage component.
type ShardCounts struct {
	Shmap     uint8 `yaml:"shmap"`
	Shhash    uint8 `yaml:"shhash"`
	Shcounter uint8 `yaml:"shcounter"`
	Shrbuffer uint8 `yaml:"shrbuffer"`
}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	return Config{
		CleanerInterval: 10,
		Namespace: NamespaceConfig{
			ShardCounts: ShardCounts{
				Shmap:     8,
				Shhash:    8,
				Shcounter: 8,
				Shrbuffer: 8,
			},
		},
	}
}
