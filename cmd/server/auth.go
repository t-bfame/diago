package server

import (
	"net/url"

	kratos "github.com/ory/kratos-client-go/client"
	"github.com/ory/kratos-client-go/client/health"
	"github.com/ory/kratos-client-go/client/version"
	log "github.com/sirupsen/logrus"
)

func NewKratosClient() (*kratos.OryKratos, error) {
	// Switch to public url? (since we're using the public whoami endpoint)
	adminURL, err := url.Parse("http://localhost:4434")

	if err != nil {
		log.Error(err)
		return nil, err
	}

	admin := kratos.NewHTTPClientWithConfig(
		nil,
		&kratos.TransportConfig{
			Schemes:  []string{adminURL.Scheme},
			Host:     adminURL.Host,
			BasePath: adminURL.Path})

	vers, err := admin.Version.GetVersion(version.NewGetVersionParams())
	if err != nil {
		log.Error(err)
		return nil, err
	}
	log.Info("Version: " + vers.GetPayload().Version)

	ok, err := admin.Health.IsInstanceAlive(health.NewIsInstanceAliveParams())
	if err != nil {
		log.Error(err)
		return nil, err
	}
	log.Info("Status: " + ok.Payload.Status)

	return admin, nil
}
