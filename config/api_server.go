package config
type ApiServer struct{
	ApiPort int  `mapstructure:"api_port" json:"api_port" yaml:"api_port"`
	ApiReadTimeout int `mapstructure:"api_read_timeout" json:"api_read_timeout" yaml:"api_read_timeout"`
	ApiWriteTimeout int `mapstructure:"api_write_timeout" json:"api_write_timeout" yaml:"api_write_timeout"`
	ApiWebRoot string `mapstructure:"api_web_root" json:"api_web_root" yaml:"api_web_root"`
}
