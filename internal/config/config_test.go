// Package config contains functions for working with application configuration.
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
			tt.c.SetBaseURL("")
			tt.c.SetBufferSize(0)
			tt.c.SetClosingTimeout(0)
			tt.c.SetTickerDuration(0)
			tt.c.GetAddress()
			tt.c.GetBaseURL()
			tt.c.GetBufferSize()
			tt.c.GetClosingTimeout()
			tt.c.GetDatabaseDSN()
			tt.c.GetFilePath()
			tt.c.GetSecretKey()
			tt.c.GetTickerDuration()
			t.Log("Supadupatest done")
		})
	}
}
