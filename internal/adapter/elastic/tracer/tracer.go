package tracer

import (
	"fmt"
	"net/url"

	"go.elastic.co/apm/v2"
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

func InitTracer(cfg *Config) (Tracer, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}

	if cfg.ServiceName == "" || cfg.ServerURL == "" || cfg.ServiceVersion == "" {
		return nil, fmt.Errorf("service name, server URL, and service version are required")
	}

	parsedURL, err := url.Parse(cfg.ServerURL)
	if err != nil {
		return nil, fmt.Errorf("invalid server URL: %w", err)
	}

	httpTransport, err := transport.NewHTTPTransport(
		transport.HTTPTransportOptions{
			ServerURLs:  []*url.URL{parsedURL},
			SecretToken: cfg.SecretToken,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create transport: %w", err)
	}

	t, err := apm.NewTracerOptions(
		apm.TracerOptions{
			ServiceName:        cfg.ServiceName,
			ServiceVersion:     cfg.ServiceVersion,
			ServiceEnvironment: cfg.Environment,
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
	if t.tracer != nil {
		return t.tracer
	}

	return nil
}

func (t *apmTracer) Shutdown() {
	if t.tracer != nil {
		t.tracer.Close()
	}
}
