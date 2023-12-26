package config

type Params struct {
	Port string `decs:"Порт запуска HTTP-сервера"`
	ResURL string `decs:"Базовый адрес результирующего сокращённого URL"`
}