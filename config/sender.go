package config

type Sender struct {
	PushGatewayUrl string   `mapstructure:"pushgateway_url" json:"pushgateway_url" yaml:"pushgateway_url"`
	WebhookUrl     string   `mapstructure:"webhook_url" json:"webhook_url" yaml:"webhook_url"`
	JobName        string   `mapstructure:"job_name" json:"job_name" yaml:"job_name"`
	ZabbixHost     string   `mapstructure:"zabbix_host" json:"zabbix_host" yaml:"zabbix_host"`
	ZabbixPort     int      `mapstructure:"zabbix_port" json:"zabbix_port" yaml:"zabbix_port"`
	ZbxItmeKey     string   `mapstructure:"zbx_itme_key" json:"zbx_itme_key" yaml:"zbx_itme_key"`
	SenderDir      string   `mapstructure:"sender_dir" json:"sender_dir" yaml:"sender_dir"`
	Senders        []string `mapstructure:"senders" json:"senders" yaml:"senders"`
}
