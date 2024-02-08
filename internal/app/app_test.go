package app

import (
	"context"
	"testing"

	"github.com/keycubasic/shortener/internal/config"
	"github.com/keycubasic/shortener/internal/storage"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateShortLink(t *testing.T) {
	type args struct {
		link string
	}

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "TestCreateShortLinkOK",
			args:    args{"https://practicum.yandex.ru"},
			want:    "a9tbDian",
			wantErr: false,
		},
	}

	ctx := context.Background()
	cfg := config.NewCfg()
	conn, err := storage.NewLinks(ctx, cfg)
	if err != nil {
		log.Fatal(err.Error())
	}
	app := NewApp(conn)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := app.CreateShortLink(ctx, tt.args.link)

			if !tt.wantErr {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Equal(t, got, tt.want, "CreateShortLink() = %v, want %v", got, tt.want)
			}
		})
	}
}
