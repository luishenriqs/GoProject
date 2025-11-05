package configs

import (
	"github.com/go-chi/jwtauth"
	"github.com/spf13/viper"
)

var cfg *config

type config struct {
	DBDriver      string           `mapstructure:"DB_DRIVER"`
	DBHost        string           `mapstructure:"DB_HOST"`
	DBPort        string           `mapstructure:"DB_PORT"`
	DBUser        string           `mapstructure:"DB_USER"`
	DBPassword    string           `mapstructure:"DB_PASSWORD"`
	DBName        string           `mapstructure:"DB_NAME"`
	WebServerPort string           `mapstructure:"WEB_SERVER_PORT"`
	JWTSecret     string           `mapstructure:"JWT_SECRET"`
	JWTExpiresIn  int              `mapstructure:"JWT_EXPIRES_IN"`
	TokenAuth     *jwtauth.JWTAuth `mapstructure:"_"`
}

func LoadConfig(path string) (*config, error) {
	v := viper.New() // Viper é criado

	// Carrega o .env localizado em `path`
	v.AddConfigPath(path)
	v.SetConfigFile(".env") // Equivale a SetConfigName(".env"), mas explícito
	if err := v.ReadInConfig(); err != nil {
		panic(err)
	}

	// AutomaticEnv() ativa a leitura das variáveis de ambiente do SO.
	// Regra de precedência: variável de ambiente sobrepõe o valor do .env
	// quando o nome da chave coincide (ex.: DB_DRIVER do ambiente vence o DB_DRIVER do arquivo).

	v.AutomaticEnv()

	var c config

	// Unmarshal(&c) preenche a sua struct config usando os mapstructure tags como “nomes das chaves”.
	if err := v.Unmarshal(&c); err != nil {
		panic(err)
	}

	// Inicializa o TokenAuth com o segredo lido
	c.TokenAuth = jwtauth.New("HS256", []byte(c.JWTSecret), nil)

	// Atualiza o singleton (mantendo seu padrão)
	cfg = &c
	return cfg, nil
}
