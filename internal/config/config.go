// Package config contains functions for working with application configuration.
package config

import (
	_ "embed"
	"flag"
	"os"

	jsoniter "github.com/json-iterator/go"
)

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary
	//go:embed config.json
	embedConfig []byte
)

// Cfg stores the application configuration.
type Cfg struct {
	ServerAddress  string `json:"server_address"`
	BaseURL        string `json:"base_url"`
	FilePath       string `json:"file_storage_path"`
	DatabaseDSN    string `json:"database_dsn"`
	SecretKey      string `json:"-"`
	BufferSize     int    `json:"-"`
	TickerDuration int    `json:"-"`
	ClosingTimeout int    `json:"-"`
	EnableHTTPS    bool   `json:"enable_https"`
	TrustedSubnet  string `json:"trusted_subnet"`
}

// ParseCfg parses the application configuration.
func (c *Cfg) ParseCfg() {
	if embedConfig != nil && json.Valid(embedConfig) {
		json.Unmarshal(embedConfig, c)
	}

	if flag.Lookup("a") == nil {
		flag.StringVar(&c.ServerAddress, "a", c.ServerAddress, "HTTP server start address")
	}

	if flag.Lookup("b") == nil {
		flag.StringVar(&c.BaseURL, "b", c.BaseURL, "Base address of the resulting shortened URL")
	}

	if flag.Lookup("f") == nil {
		flag.StringVar(&c.FilePath, "f", c.FilePath, "Path to disk storage")
	}

	if flag.Lookup("d") == nil {
		flag.StringVar(&c.DatabaseDSN, "d", c.DatabaseDSN, "Database connection address")
	}

	if flag.Lookup("k") == nil {
		flag.StringVar(&c.SecretKey, "k", "keycubasic", "JWT secret key")
	}

	if flag.Lookup("w") == nil {
		flag.IntVar(&c.BufferSize, "w", 256, "Size of worker buffer")
	}

	if flag.Lookup("u") == nil {
		flag.IntVar(&c.TickerDuration, "u", 10, "Deletion ticker duration (seconds)")
	}

	if flag.Lookup("c") == nil {
		flag.IntVar(&c.ClosingTimeout, "c", 10, "Termination timeout (seconds)")
	}

	if flag.Lookup("s") == nil {
		flag.BoolFunc("s", "Sign of https protocol support", func(s string) error {
			c.EnableHTTPS = true
			return nil
		})
	}

	if flag.Lookup("t") == nil {
		flag.StringVar(&c.TrustedSubnet, "t", c.TrustedSubnet, "Trusted subnet")
	}

	flag.Parse()

	if _, ok := os.LookupEnv("ENABLE_HTTPS"); ok {
		c.EnableHTTPS = ok
	}

	if a := os.Getenv("SERVER_ADDRESS"); a != "" {
		c.ServerAddress = a
	}

	if b := os.Getenv("BASE_URL"); b != "" {
		c.BaseURL = b
	}

	if f := os.Getenv("FILE_STORAGE_PATH"); f != "" {
		c.FilePath = f
	}

	if d := os.Getenv("DATABASE_DSN"); d != "" {
		c.DatabaseDSN = d
	}

	if k := os.Getenv("JWT_SECRET_KEY"); k != "" {
		c.SecretKey = k
	}

	if t := os.Getenv("TRUSTED_SUBNET"); t != "" {
		c.TrustedSubnet = t
	}
}
