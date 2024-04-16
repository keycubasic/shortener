package app

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/keycubasic/shortener/internal/config"
	"github.com/keycubasic/shortener/internal/storage"
	"github.com/keycubasic/shortener/internal/storage/mock"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetLinkMap(t *testing.T) {
	link := "https://practicum.yandex.ru"
	ctx := context.Background()
	cfg := config.NewCfg()
	db, err := storage.NewLinks(ctx, cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

	app := NewApp(ctx, cfg, db)
	shortLink, err := app.CreateShortLink(ctx, 1, link)
	if err != nil {
		log.Fatal(err.Error())
	}

	tests := []struct {
		name      string
		shortLink string
		want      string
		wantErr   bool
	}{
		{
			name:      "TestGetLinkMapOK",
			shortLink: shortLink,
			want:      link,
			wantErr:   false,
		},
		{
			name:      "TestGetLinkMapKO",
			shortLink: "wrongShortLink",
			want:      "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := app.GetLink(ctx, tt.shortLink)
			if (err != nil) != tt.wantErr {
				t.Errorf("App.GetLink() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("App.GetLink() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateShortLinkMap(t *testing.T) {
	ctx := context.Background()
	cfg := config.NewCfg()
	db, err := storage.NewLinks(ctx, cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

	app := NewApp(ctx, cfg, db)

	tests := []struct {
		name    string
		link    string
		want    string
		wantErr bool
	}{
		{
			name:    "TestCreateShortLinkMapOK",
			link:    "https://practicum.yandex.ru",
			want:    "a9tbDian",
			wantErr: false,
		},
		{
			name:    "TestCreateShortLinkMapKO",
			link:    "https://practicum.yandex.ru",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := app.CreateShortLink(ctx, 0, tt.link)

			if !tt.wantErr {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Equal(t, got, tt.want, "CreateShortLink() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPingDB(t *testing.T) {
	ctx := context.Background()
	cfg := config.NewCfg()

	ctrlDB := gomock.NewController(t)
	defer ctrlDB.Finish()

	mockDB := mock.NewMockDatabase(ctrlDB)
	mockDB.EXPECT().PingContext(gomock.Any()).Return(nil)
	mockDB.EXPECT().PingContext(gomock.Any()).Return(errors.New("oops!"))

	app := NewApp(ctx, cfg, mockDB)

	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "TestPingDBOK",
			wantErr: false,
		},
		{
			name:    "TestPingDBKO",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := app.Ping(ctx); (err != nil) != tt.wantErr {
				t.Errorf("App.Ping() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetShortLinkDB(t *testing.T) {
	link := "https://practicum.yandex.ru"
	ctx := context.Background()
	cfg := config.NewCfg()

	ctrlDB := gomock.NewController(t)
	defer ctrlDB.Finish()

	mockDB := mock.NewMockDatabase(ctrlDB)
	mockDB.EXPECT().GetShortLink(gomock.Any(), gomock.Any()).Return(gomock.Any().String(), nil)
	mockDB.EXPECT().GetShortLink(gomock.Any(), gomock.Any()).Return("", errors.New("oops!"))

	app := NewApp(ctx, cfg, mockDB)

	tests := []struct {
		name    string
		link    string
		want    string
		wantErr bool
	}{
		{
			name:    "TestGetShortLinkDBOK",
			link:    link,
			want:    gomock.Any().String(),
			wantErr: false,
		},
		{
			name:    "TestGetShortLinkDBKO",
			link:    link,
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := app.GetShortLink(ctx, tt.link)
			if (err != nil) != tt.wantErr {
				t.Errorf("App.GetShortLink() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("App.GetShortLink() = %v, want %v", got, tt.want)
			}
		})
	}
}
