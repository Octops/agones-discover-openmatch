package config

import (
	"github.com/kelseyhightower/envconfig"
	"log"
)

type OpenMatchConnConfig struct {
	FrontEnd     string `envconfig:"frontend_addr"`
	BackEnd      string `envconfig:"backend_addr"`
	QueryService string `envconfig:"query_service_addr"`
}

func OpenMatch() OpenMatchConnConfig {
	var config OpenMatchConnConfig
	err := envconfig.Process("openmatch", &config)
	if err != nil {
		log.Fatal(err.Error())
	}

	return config
}
