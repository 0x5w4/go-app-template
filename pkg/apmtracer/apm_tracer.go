package apmtracer

import (
	"errors"
	"fmt"
	"net/url"

	apm "go.elastic.co/apm/v2"
	"go.elastic.co/apm/v2/transport"
)

type Tracer interface {
	Shutdown()
	Tracer() *apm.Tracer
}

type apmTracer struct {
	tracer *apm.Tracer
}

type Config struct {
	ServiceName    string
	ServiceVersion string
	ServerURL      string
	SecretToken    string
	Environment    string
	NodeName       string
}

func NewApmTracer(config *Config) (Tracer, error) {
	if config == nil {
		return nil, errors.New("config is required")
	}

	if config.ServiceName == "" || config.ServerURL == "" || config.ServiceVersion == "" {
		return nil, errors.New("service name, server URL, and service version are required")
	}

	parsedURL, err := url.Parse(config.ServerURL)
	if err != nil {
		return nil, fmt.Errorf("invalid server URL: %w", err)
	}

	httpTransport, err := transport.NewHTTPTransport(
		transport.HTTPTransportOptions{
			ServerURLs:  []*url.URL{parsedURL},
			SecretToken: config.SecretToken,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create transport: %w", err)
	}

	t, err := apm.NewTracerOptions(
		apm.TracerOptions{
			ServiceName:        config.ServiceName,
			ServiceVersion:     config.ServiceVersion,
			ServiceEnvironment: config.Environment,
			Transport:          httpTransport,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create elastic apm tracer: %w", err)
	}

	apm.SetDefaultTracer(t)

	return &apmTracer{
		tracer: t,
	}, nil
}

func (t *apmTracer) Tracer() *apm.Tracer {
	return t.tracer
}

func (t *apmTracer) Shutdown() {
	if t.tracer != nil {
		t.tracer.Close()
	}
}
