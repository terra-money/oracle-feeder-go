package configs

type Config struct {
	Port             int                       `json:"port,omitempty"`
	MetricsPort      int                       `json:"metrics_port,omitempty"`
	Sentry           string                    `json:"sentry,omitempty"` // sentry dsn (https://sentry.io/ - error reporting service)
	Providers        map[string]ProviderConfig `json:"providers,omitempty"`
	ProviderPriority []string                  `json:"provider_prioirty,omitempty"`
}

type ProviderConfig struct {
	Symbols  []string `json:"symbols,omitempty"`
	Interval int      `json:"interval,omitempty"` // in seconds
	Timeout  int      `json:"timeout,omitempty"`
}
