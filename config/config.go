package config

// Config represents configuration for the memap core.
type Config struct {
	CleanerInterval int             `yaml:"cleanerInterval" json:"cleanerInterval"`
	Namespace       NamespaceConfig `yaml:"namespace" json:"namespace"`
}

// NamespaceConfig represents configuration for namespaces.
type NamespaceConfig struct {
	ShardCounts ShardCounts `yaml:"shardCounts" json:"shardCounts"`
}

// ShardCounts specifies the number of shards for each storage component.
type ShardCounts struct {
	Shmap     uint8 `yaml:"shmap" json:"shmap"`
	Shhash    uint8 `yaml:"shhash" json:"shhash"`
	Shcounter uint8 `yaml:"shcounter" json:"shcounter"`
	Shrbuffer uint8 `yaml:"shrbuffer" json:"shrbuffer"`
	Shset     uint8 `yaml:"shset" json:"shset"`
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
				Shset:     8,
			},
		},
	}
}
