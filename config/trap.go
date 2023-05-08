package config

type TrapServer struct {
	Ip              string `mapstructure:"ip" json:"ip" yaml:"ip"`
	Port            int64  `mapstructure:"port" json:"port" yaml:"port"`
	Version         string `mapstructure:"version" json:"version" yaml:"version"`
	Community       string `mapstructure:"community" json:"community" yaml:"community"`
	ReadCommunity   string `mapstructure:"read_community" json:"read_community" yaml:"read_community"`
	Timeout         int64  `mapstructure:"timeout" json:"timeout" yaml:"timeout"`
	Maxoids         int64  `mapstructure:"maxoids" json:"maxoids" yaml:"maxoids"`
	MibMapFile      string `mapstructure:"mib_map_file" json:"mib_map_file" yaml:"mib_map_file"`
	BlackMibMapFile string `mapstructure:"black_mib_map_file" json:"black_mib_map_file" yaml:"black_mib_map_file"`
}
