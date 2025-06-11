package config

type Config struct {
	Localhost string
	Port      string
}

func NewConfig() *Config {
	return &Config{
		Localhost: "localhost",
		Port:      "8080",
	}
}
