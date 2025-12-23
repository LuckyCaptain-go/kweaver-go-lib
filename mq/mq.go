package mq

type MQAuthSetting struct {
	Username  string
	Password  string `json:"-"`
	Mechanism string
}

// mq配置项
type MQSetting struct {
	MQType string
	MQHost string
	MQPort int
	Tenant string
	Auth   MQAuthSetting
}
