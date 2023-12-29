package config

import (
	"flag"
	"os"
)

type Cfg struct {
	address string
	baseUrl string
}

func NewCfg() (*Cfg, error) {
	cfg := new(Cfg)
	return cfg, nil
}

func (c *Cfg) GetAddress() string {
	return c.address
}

func (c *Cfg) SetAddress(address string) {
	c.address = address
}

func (c *Cfg) GetBaseUrl() string {
	return c.baseUrl
}

func (c *Cfg) SetBaseUrl(baseUrl string) {
	c.baseUrl = baseUrl
}

func (c *Cfg) ParseCfg() {
	cfg := new(Cfg)
	flag.StringVar(&cfg.address, "a", ":8080", "HTTP server start address")
	flag.StringVar(&cfg.baseUrl, "b", "http://localhost:8080", "Base address of the resulting shortened URL")
	flag.Parse()
	if a := os.Getenv("SERVER_ADDRESS"); a != "" {
		cfg.address = a
	}
	if b := os.Getenv("BASE_URL"); b != "" {
		cfg.baseUrl = b
	}
}
