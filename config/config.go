package config

type Server struct {
	LogConf    `mapstructure:"logconf" json:"logconf" yaml:"logconf"`
	TrapServer `mapstructure:"trapserver" json:"trapserver" yaml:"trapserver"`
	ApiServer  `mapstructure:"apiserver" json:"apiserver" yaml:"apiserver"`
	Sender     `mapstructure:"sender" json:"sender" yaml:"sender"`
}
