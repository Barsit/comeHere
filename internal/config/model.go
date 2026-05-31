package config

import "time"

type HijackRule struct {
	ID          string    `json:"id"`
	Source      string    `json:"source"`
	SourcePort  int       `json:"source_port"`
	Target      string    `json:"target"`
	TargetTLS   bool      `json:"target_tls"`
	TargetHost  string    `json:"target_host"`
	Description string    `json:"description"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Config struct {
	Rules      []HijackRule `json:"rules"`
	ListenPort int          `json:"listen_port"`
	AdminPort  int          `json:"admin_port"`
	CADays     int          `json:"ca_days"`
}

func DefaultConfig() *Config {
	return &Config{
		Rules:      []HijackRule{},
		ListenPort: 443,
		AdminPort:  8848,
		CADays:     3650,
	}
}
