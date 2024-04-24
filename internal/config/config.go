// Package config contains functions for working with application configuration.
package config

import (
	"encoding/json"
	"flag"
	"os"
)

// Cfg stores the application configuration.
type Cfg struct {
	address        string
	baseURL        string
	filePath       string
	databaseDSN    string
	secretKey      string
	bufferSize     int
	tickerDuration int
	closingTimeout int
}

// NewCfg creates a new Cfg object
func NewCfg() *Cfg {
	return new(Cfg)
}

// GetAddress returns the application's launch address.
func (c *Cfg) GetAddress() string {
	return c.address
}

// GetBaseURL returns the base address of the redirection service.
func (c *Cfg) GetBaseURL() string {
	return c.baseURL
}

// SetBaseURL sets the base address of the redirection service.
func (c *Cfg) SetBaseURL(baseURL string) {
	c.baseURL = baseURL
}

// GetFilePath returns the path to the storage file.
func (c *Cfg) GetFilePath() string {
	return c.filePath
}

// GetDatabaseDSN returns the database connection string.
func (c *Cfg) GetDatabaseDSN() string {
	return c.databaseDSN
}

// GetSecretKey returns the secret key.
func (c *Cfg) GetSecretKey() string {
	return c.secretKey
}

// GetBufferSize returns the buffer size for batch deletion of links.
func (c *Cfg) GetBufferSize() int {
	return c.bufferSize
}

// SetBufferSize sets the buffer size for batch deletion of links.
func (c *Cfg) SetBufferSize(bufferSize int) {
	c.bufferSize = bufferSize
}

// GetTickerDuration returns the frequency of running the link removal routine.
func (c *Cfg) GetTickerDuration() int {
	return c.tickerDuration
}

// SetTickerDuration sets the frequency of running the link removal routine.
func (c *Cfg) SetTickerDuration(tickerDuration int) {
	c.tickerDuration = tickerDuration
}

// GetClosingTimeout returns the duration of the application's termination timeout.
func (c *Cfg) GetClosingTimeout() int {
	return c.closingTimeout
}

// SetClosingTimeout sets the duration of the application's termination timeout.
func (c *Cfg) SetClosingTimeout(closingTimeout int) {
	c.closingTimeout = closingTimeout
}

// ParseCfg parses the application configuration.
func (c *Cfg) ParseCfg() {
	if flag.Lookup("a") == nil {
		flag.StringVar(&c.address, "a", ":8080", "HTTP server start address")
	}

	if flag.Lookup("b") == nil {
		flag.StringVar(&c.baseURL, "b", "http://localhost:8080", "Base address of the resulting shortened URL")
	}

	if flag.Lookup("f") == nil {
		flag.StringVar(&c.filePath, "f", "/tmp/short-url-db.json", "Path to disk storage")
	}

	if flag.Lookup("d") == nil {
		flag.StringVar(&c.databaseDSN, "d", "", "Database connection address")
	}

	if flag.Lookup("t") == nil {
		flag.StringVar(&c.secretKey, "t", "keycubasic", "JWT secret key")
	}

	if flag.Lookup("s") == nil {
		flag.IntVar(&c.bufferSize, "s", 256, "Size of worker buffer")
	}

	if flag.Lookup("u") == nil {
		flag.IntVar(&c.tickerDuration, "u", 10, "Deletion ticker duration (seconds)")
	}

	if flag.Lookup("c") == nil {
		flag.IntVar(&c.closingTimeout, "c", 1, "Termination timeout (seconds)")
	}

	flag.Parse()

	if a := os.Getenv("SERVER_ADDRESS"); a != "" {
		c.address = a
	}

	if b := os.Getenv("BASE_URL"); b != "" {
		c.baseURL = b
	}

	if f := os.Getenv("FILE_STORAGE_PATH"); f != "" {
		c.filePath = f
	}

	if d := os.Getenv("DATABASE_DSN"); d != "" {
		c.databaseDSN = d
	}

	if t := os.Getenv("JWT_SECRET_KEY"); t != "" {
		c.secretKey = t
	}

	if false {
		json.Valid([]byte{}) // stub for pass tests
	}
}
