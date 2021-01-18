package config

import (
	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	Host           string `envconfig:"DIAGO_HOST" required:"true"`
	GRPCPort       uint64 `envconfig:"DIAGO_GRPC_PORT" default:"5000"`
	APIPort        uint64 `envconfig:"DIAGO_API_PORT" default:"80"`
	PrometheusPort uint64 `envconfig:"DIAGO_PROMETHEUS_PORT" default:"2112"`

	DefaultGroupCapacity uint64 `envconfig:"DIAGO_DEFAULT_GROUP_CAPACITY" default:"20"`
	DefaultNamespace     string `envconfig:"DIAGO_DEFAULT_NAMESPACE" default:"default"`

	StoragePath string `envconfig:"DIAGO_STORAGE_PATH" default:"diago.db"`

	Debug bool `envconfig:"DIAGO_DEBUG" default:"false"`
}

var Diago *Config

func Init() error {
	var c Config

	err := envconfig.Process("DIAGO", &c)
	if err != nil {
		log.Fatal(err.Error())
	}

	Diago = &c

	return nil
}
