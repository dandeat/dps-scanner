package models

type EnvironmentConfig struct {
	AppEnv     string `json:"app_env"`
	AppName    string `json:"app_name"`
	AppVersion string `json:"app_version"`
}

type Config struct {
	DriverName string   `json:"driver"`
	Database   Database `json:"database"`
	Redis      Redis    `json:"redis"`
}

type Database struct {
	DBName  string `json:"db_name"`
	DBPort  string `json:"db_port"`
	DBUser  string `json:"db_user"`
	DBPass  string `json:"db_pass"`
	SSLMode string `json:"ssl_mode"`
}

type Redis struct {
	REDISAddr string `json:"redis_addr"`
	REDISPass string `json:"redis_pass"`
}

type RabbitMq struct {
	USERRMQ string `json:"user_rmq"`
	PASSRMQ string `json:"pass_rmq"`
	HOSTRMQ string `json:"host_rmq"`
	PORTRMQ string `json:"port_rmq"`
	VHOSTMQ string `json:"vhost_rmq"`
}

type ReqHeader struct {
	Header []Header
}

type Header struct {
	Key      string
	Val      string
	IsUpCase bool
}
