package utils

// config
type AppConfig struct {
	DbFilename string `env:"WALI_DB" env-default:"wali.db"`

	WorkersCount int `env:"WALI_WORKERS_COUNT" env-default:"3"`
}
