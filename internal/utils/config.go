package utils

// config
type AppConfig struct {
	DbFilename string `env:"WALI_DB" env-default:"wali.db"`

	WorkersCount  int `env:"WALI_WORKERS_COUNT" env-default:"3"`
	ShowTickEvery int `env:"WALI_SHOW_TICK_EVERY" env-default:"10"`
}
