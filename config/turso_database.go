package config

type Turso struct {
	DBName string `toml:"DBName" env:"TURSO_DBNAME" env-default:"turso"`
	URL    string `toml:"URL" env:"TURSO_URL" env-default:"http://localhost:8080"`
}
