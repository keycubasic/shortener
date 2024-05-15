// Package config contains functions for working with application configuration.
package config

import (
	_ "embed"
	"flag"
	"os"
	"time"

	jsoniter "github.com/json-iterator/go"
)

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary
	//go:embed config.json
	embedConfig []byte
)

// Cfg stores the application configuration.
type Cfg struct {
	ServerAddress  string        `json:"server_address"`
	BaseURL        string        `json:"base_url"`
	FilePath       string        `json:"file_storage_path"`
	DatabaseDSN    string        `json:"database_dsn"`
	SecretKey      string        `json:"-"`
	BufferSize     int           `json:"-"`
	TickerDuration time.Duration `json:"-"`
	ClosingTimeout time.Duration `json:"-"`
	EnableHTTPS    bool          `json:"enable_https"`
}

// ParseCfg parses the application configuration.
func (c *Cfg) ParseCfg() {
	/*if embedConfig != nil && json.Valid(embedConfig) {
		json.Unmarshal(embedConfig, c)
	}*/

	if flag.Lookup("a") == nil {
		flag.StringVar(&c.ServerAddress, "a", ":8080", "HTTP server start address")
	}

	if flag.Lookup("b") == nil {
		flag.StringVar(&c.BaseURL, "b", "http://localhost:8080", "Base address of the resulting shortened URL")
	}

	if flag.Lookup("f") == nil {
		flag.StringVar(&c.FilePath, "f", "/tmp/short-url-db.json", "Path to disk storage")
	}

	if flag.Lookup("d") == nil {
		flag.StringVar(&c.DatabaseDSN, "d", "", "Database connection address")
	}

	if flag.Lookup("t") == nil {
		flag.StringVar(&c.SecretKey, "t", "keycubasic", "JWT secret key")
	}

	if flag.Lookup("w") == nil {
		flag.IntVar(&c.BufferSize, "w", 256, "Size of worker buffer")
	}

	if flag.Lookup("u") == nil {
		flag.DurationVar(&c.TickerDuration, "u", 10, "Deletion ticker duration (seconds)")
	}

	if flag.Lookup("c") == nil {
		flag.DurationVar(&c.ClosingTimeout, "c", 1, "Termination timeout (seconds)")
	}

	/*if flag.Lookup("s") == nil {
		flag.BoolFunc("s", "Sign of https protocol support", func(s string) error {
			c.EnableHTTPS = true
			return nil
		})
	}*/

	flag.Parse()

	/*if _, ok := os.LookupEnv("ENABLE_HTTPS"); ok {
		c.EnableHTTPS = ok
	}*/

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

	if t := os.Getenv("JWT_SECRET_KEY"); t != "" {
		c.SecretKey = t
	}
}
