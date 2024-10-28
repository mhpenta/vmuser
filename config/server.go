package config

type Server struct {
	Port string `toml:"Port" env:"SERVER_PORT" env-default:"10101"`
}
