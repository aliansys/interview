package config

type Config struct {
	Api struct {
		Address string `yaml:"address"`
	} `yaml:"api"`
	ClickHouse struct {
		DSN       string `yaml:"dsn"`
		BatchSize int    `yaml:"batch_size"`
	} `yaml:"clickhouse"`
}
