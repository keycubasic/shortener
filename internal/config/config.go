package config

import (
	"flag"
	"os"
)

type Cfg struct {
	address     string
	baseURL     string
	filePath    string
	databaseDSN string
	secretKey   string
}

func NewCfg() *Cfg {
	return new(Cfg)
}

func (c *Cfg) GetAddress() string {
	return c.address
}

func (c *Cfg) SetAddress(address string) {
	c.address = address
}

func (c *Cfg) GetBaseURL() string {
	return c.baseURL
}

func (c *Cfg) SetBaseURL(baseURL string) {
	c.baseURL = baseURL
}

func (c *Cfg) GetFilePath() string {
	return c.filePath
}

func (c *Cfg) SetFilePath(filePath string) {
	c.filePath = filePath
}

func (c *Cfg) GetDatabaseDSN() string {
	return c.databaseDSN
}

func (c *Cfg) SetDatabaseDSN(databaseDSN string) {
	c.databaseDSN = databaseDSN
}

func (c *Cfg) GetSecretKey() string {
	return c.secretKey
}

func (c *Cfg) SetSecretKey(secretKey string) {
	c.secretKey = secretKey
}

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
		flag.StringVar(&c.secretKey, "t", "keycubasic", "JWT secret key ")
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
}
