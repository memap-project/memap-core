package config

type Config struct {
	CleanerInterval int             `yaml:"cleanerInterval"`
	Namespace       NamespaceConfig `yaml:"namespace"`
}

type NamespaceConfig struct {
	ShardCounts ShardCounts `yaml:"shardCounts"`
}

type ShardCounts struct {
	Shmap     uint8 `yaml:"shmap"`
	Shhash    uint8 `yaml:"shhash"`
	Shcounter uint8 `yaml:"shcounter"`
	Shrbuffer uint8 `yaml:"shrbuffer"`
}

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
