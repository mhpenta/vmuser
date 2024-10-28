package config

type Elastic struct {
	Addresses string `toml:"Addresses" env:"ELASTIC_ADDRESSES" env-default:"https://localhost:9200"`
	Username  string `toml:"Username" env:"ELASTIC_USERNAME" env-default:"elastic"`
	Password  string `toml:"Password" env:"ELASTIC_PASSWORD"`
}
