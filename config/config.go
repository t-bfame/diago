package config

import (
	"github.com/gobuffalo/packr/v2"
	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	Host           string `envconfig:"DIAGO_HOST" required:"true"`
	GRPCPort       uint64 `envconfig:"DIAGO_GRPC_PORT" default:"5000"`
	APIPort        uint64 `envconfig:"DIAGO_API_PORT" default:"80"`
	MongoDBHost    string `envconfig:"MONGO_DB_HOST" required:"true"`
	MongoDBPort    uint64 `envconfig:"MONGO_DB_PORT" default:"27017"`
	PrometheusPort uint64 `envconfig:"DIAGO_PROMETHEUS_PORT" default:"2112"`

	DefaultGroupCapacity uint64 `envconfig:"DIAGO_DEFAULT_GROUP_CAPACITY" default:"200"`
	DefaultNamespace     string `envconfig:"DIAGO_DEFAULT_NAMESPACE" default:"default"`

	StoragePath string `envconfig:"DIAGO_STORAGE_PATH" default:"diago.db"`

	Debug bool `envconfig:"DIAGO_DEBUG" default:"false"`

	GrafanaBasePath        string `envconfig:"DIAGO_GRAFANA_BASE_PATH" default:""`
	GrafanaAPIKey          string `envconfig:"DIAGO_GRAFANA_API_KEY" default:""`
	GrafanaDashboardConfig string `envconfig:"DIAGO_GRAFANA_DASHBOARD_CONFIG"`
}

var Diago *Config

// Initializes a new Diago config
// For GrafanaDashboardConfig, if the env variable is not set then defaults
// to the default grafana-dash.json from packr2
func Init() error {
	var c Config

	err := envconfig.Process("DIAGO", &c)
	if err != nil {
		log.Fatal(err.Error())
	}

	Diago = &c

	if Diago.GrafanaDashboardConfig == "" {
		box := packr.New("grafana", "../static")
		config, _ := box.FindString("grafana-dash.json")
		c.GrafanaDashboardConfig = string(config)
	}

	return nil
}
