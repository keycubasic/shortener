package config

import (
	"flag"
	"log"
	"os"

	// "github.com/caarlos0/env/v6"
	"github.com/octago/sflags/gen/gflag"
)

type config struct {
	A string `decs:"Порт запуска HTTP-сервера" env:"SERVER_ADDRESS"`
	B string `decs:"Базовый адрес результирующего сокращённого URL" env:"BASE_URL"`
}

var Config = &config{
	A: ":8080",
	B: "http://localhost:8080",
}

func ParseFlags() {
	if err := gflag.ParseToDef(Config); err != nil {
		log.Fatalln(err)
	}
	flag.Parse()
}

func ParseConfig() {
	ParseFlags()
	if a := os.Getenv("SERVER_ADDRESS"); a != "" {
		Config.A = a
	}
	if b := os.Getenv("BASE_URL"); b != "" {
		Config.B = b
	}
}
