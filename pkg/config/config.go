package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the root configuration for the push proxy server
type Config struct {
	Server ServerConfig `yaml:"server"`
	Log    LogConfig    `yaml:"log"`
	FCM    FCMConfig    `yaml:"fcm"`
	APNS   APNSConfig   `yaml:"apns"`
}

// ServerConfig holds HTTP server options
type ServerConfig struct {
	ListenAddress string `yaml:"listen_address"` // e.g. "0.0.0.0:8066"
	ReadTimeout   int    `yaml:"read_timeout"`   // in seconds
	WriteTimeout  int    `yaml:"write_timeout"`  // in seconds
	MaxHeaderBytes int   `yaml:"max_header_bytes"`
	EnableRateLimit bool `yaml:"enable_rate_limit"`
	RateLimitPerSec float64 `yaml:"rate_limit_per_sec"`
	RateLimitBurst int   `yaml:"rate_limit_burst"`
	TLSCertFile   string `yaml:"tls_cert_file"`
	TLSKeyFile    string `yaml:"tls_key_file"`
}

// LogConfig holds logging options
type LogConfig struct {
	Level      string `yaml:"level"`       // "debug", "info", "warn", "error"
	Format     string `yaml:"format"`      // "text", "json"
	LogToFile  bool   `yaml:"log_to_file"` // save logs to file
	FilePath   string `yaml:"file_path"`   // e.g. "/var/log/mattermost-push-proxy.log"
}

// FCMConfig holds Firebase Cloud Messaging credentials
type FCMConfig struct {
	Enabled            bool   `yaml:"enabled"`
	CredentialsFile    string `yaml:"credentials_file"`     // Path to Google Service Account JSON
	ProjectID          string `yaml:"project_id"`           // FCM Project ID (optional if in JSON)
	AndroidRNTopic     string `yaml:"android_rn_topic"`     // Topic for Android RN app
	AndroidLegacyKey   string `yaml:"android_legacy_key"`   // Optional legacy key backup
}

// APNSConfig holds Apple Push Notification service credentials
type APNSConfig struct {
	Enabled          bool   `yaml:"enabled"`
	AuthKeyFile      string `yaml:"auth_key_file"`     // Path to .p8 key file
	KeyID            string `yaml:"key_id"`            // 10-character Key ID
	TeamID           string `yaml:"team_id"`           // 10-character Team ID
	Topic            string `yaml:"topic"`             // App Bundle Identifier (e.g. com.mattermost.rn)
	Production       bool   `yaml:"production"`        // true for production gateway, false for sandbox
	CertFile         string `yaml:"cert_file"`         // Alternative .p12 certificate file path
	CertPassword     string `yaml:"cert_password"`     // Alternative .p12 certificate password
}

// DefaultConfig returns safe defaults for production deployment
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			ListenAddress:   "0.0.0.0:8066",
			ReadTimeout:     15,
			WriteTimeout:    15,
			MaxHeaderBytes:  1 << 20, // 1MB
			EnableRateLimit: true,
			RateLimitPerSec: 100,
			RateLimitBurst:  200,
		},
		Log: LogConfig{
			Level:     "info",
			Format:    "text",
			LogToFile: false,
			FilePath:  "/var/log/mattermost-push-proxy.log",
		},
		FCM: FCMConfig{
			Enabled: true,
		},
		APNS: APNSConfig{
			Enabled:    true,
			Production: true,
		},
	}
}

// LoadConfig loads configuration from a YAML file with Environment variable overrides
func LoadConfig(path string) (*Config, error) {
	cfg := DefaultConfig()

	if path != "" {
		cleanPath := filepath.Clean(path)
		data, err := os.ReadFile(cleanPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file at %s: %w", path, err)
		}

		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse yaml config at %s: %w", path, err)
		}
	}

	// Override with environment variables if present
	overrideFromEnv(cfg)

	return cfg, nil
}

func overrideFromEnv(cfg *Config) {
	if addr := os.Getenv("MPP_LISTEN_ADDRESS"); addr != "" {
		cfg.Server.ListenAddress = addr
	}
	if port := os.Getenv("PORT"); port != "" {
		cfg.Server.ListenAddress = "0.0.0.0:" + port
	}
	if lvl := os.Getenv("MPP_LOG_LEVEL"); lvl != "" {
		cfg.Log.Level = lvl
	}
	if fcmCreds := os.Getenv("MPP_FCM_CREDENTIALS_FILE"); fcmCreds != "" {
		cfg.FCM.CredentialsFile = fcmCreds
	}
	if fcmProj := os.Getenv("MPP_FCM_PROJECT_ID"); fcmProj != "" {
		cfg.FCM.ProjectID = fcmProj
	}
	if apnsKey := os.Getenv("MPP_APNS_AUTH_KEY_FILE"); apnsKey != "" {
		cfg.APNS.AuthKeyFile = apnsKey
	}
	if apnsKeyID := os.Getenv("MPP_APNS_KEY_ID"); apnsKeyID != "" {
		cfg.APNS.KeyID = apnsKeyID
	}
	if apnsTeamID := os.Getenv("MPP_APNS_TEAM_ID"); apnsTeamID != "" {
		cfg.APNS.TeamID = apnsTeamID
	}
	if apnsTopic := os.Getenv("MPP_APNS_TOPIC"); apnsTopic != "" {
		cfg.APNS.Topic = apnsTopic
	}
	if prod := os.Getenv("MPP_APNS_PRODUCTION"); prod != "" {
		cfg.APNS.Production = strings.ToLower(prod) == "true" || prod == "1"
	}
}

// Validate checks for configuration sanity
func (c *Config) Validate() error {
	if c.Server.ListenAddress == "" {
		return fmt.Errorf("server listen_address cannot be empty")
	}
	return nil
}
