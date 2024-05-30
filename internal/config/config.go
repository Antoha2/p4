package config

type Config struct {
	DB DBConfig
}

type DBConfig struct {
	User     string
	Password string
	Host     string
	Port     int
	Dbname   string
	Sslmode  string
}

func GetConfig() *Config {

	return &Config{
		DB: DBConfig{
			Dbname:   "postgres",
			User:     "user",
			Password: "password",
			Host:     "127.0.0.1",
			//Host:    "postgres",
			Port:    5432,
			Sslmode: "",
		},
	}

}
