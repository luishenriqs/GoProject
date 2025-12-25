package configs

import (
	"github.com/go-chi/jwtauth/v5"
	"github.com/spf13/viper"
)

var cfg *Config

// Config mantém as configs carregadas do .env (ou env vars do SO).
// IMPORTANTE: precisa ser exportado para que cmd/server consiga tipar openDatabase(*configs.Config).
type Config struct {
	DBDriver      string           `mapstructure:"DB_DRIVER"`
	DBHost        string           `mapstructure:"DB_HOST"`
	DBPort        string           `mapstructure:"DB_PORT"`
	DBUser        string           `mapstructure:"DB_USER"`
	DBPassword    string           `mapstructure:"DB_PASSWORD"`
	DBName        string           `mapstructure:"DB_NAME"`
	WebServerPort string           `mapstructure:"WEB_SERVER_PORT"`
	JWTSecret     string           `mapstructure:"JWT_SECRET"`
	JWTExpiresIn  int              `mapstructure:"JWT_EXPIRESIN"`
	TokenAuth     *jwtauth.JWTAuth `mapstructure:"_"`
}

// LoadConfig carrega .env via Viper (com override por env vars do SO) e inicializa TokenAuth.
// Mantém o padrão singleton já existente.
func LoadConfig(path string) (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	v := viper.New()
	v.AddConfigPath(path)
	v.SetConfigFile(".env")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var c Config
	if err := v.Unmarshal(&c); err != nil {
		return nil, err
	}

	c.TokenAuth = jwtauth.New("HS256", []byte(c.JWTSecret), nil)

	cfg = &c
	return cfg, nil
}
