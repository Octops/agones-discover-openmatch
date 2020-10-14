package config

import (
	"github.com/kelseyhightower/envconfig"
	"log"
)

type OpenMatchConnConfig struct {
	FrontEnd string `envconfig:"frontend_endpoint"`
}

func OpenMatch() OpenMatchConnConfig {
	var config OpenMatchConnConfig
	err := envconfig.Process("openmatch", &config)
	if err != nil {
		log.Fatal(err.Error())
	}

	return config
}
