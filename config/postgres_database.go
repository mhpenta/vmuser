package config

type Postgres struct {
	Host     string `toml:"Host" env:"DB_HOST" env-default:"localhost"`
	Port     int    `toml:"Port" env:"DB_PORT" env-default:"5432"`
	User     string `toml:"User" env:"DB_User"`
	Password string `toml:"Password" env:"Password"`
	DBName   string `toml:"DBName" env:"DBName"`
	SSLMode  string `toml:"SSLMode" env:"SSLMode" env-default:"disable"`
}
