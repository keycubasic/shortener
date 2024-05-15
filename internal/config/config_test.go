// Package config contains functions for working with application *config.Cfg.
package config

import "testing"

func TestCfg(t *testing.T) {
	tests := []struct {
		name string
		c    *Cfg
	}{
		{
			name: "TextCfgOK",
			c:    new(Cfg),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.c.ParseCfg()
		})
	}
}
