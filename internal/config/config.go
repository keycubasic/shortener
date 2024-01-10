package config

import (
	"flag"
	"os"
)

type Cfg struct {
	address string
	baseURL string
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

func (c *Cfg) ParseCfg() {
	flag.StringVar(&c.address, "a", ":8080", "HTTP server start address")
	flag.StringVar(&c.baseURL, "b", "http://localhost:8080", "Base address of the resulting shortened URL")
	flag.Parse()

	if a := os.Getenv("SERVER_ADDRESS"); a != "" {
		c.address = a
	}

	if b := os.Getenv("BASE_URL"); b != "" {
		c.baseURL = b
	}
}
