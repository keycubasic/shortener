package app

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/keycubasic/shortener/internal/config"
	"github.com/keycubasic/shortener/internal/dao"
	"github.com/keycubasic/shortener/internal/dto"
	"github.com/keycubasic/shortener/internal/storage/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errTest = errors.New("oops")

func TestPing(t *testing.T) {
	ctx := context.Background()
	cfg := config.NewCfg()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := mock.NewMockDatabase(ctrl)
	mock.EXPECT().PingContext(gomock.Any()).Return(nil)
	mock.EXPECT().PingContext(gomock.Any()).Return(errTest)

	app := NewApp(ctx, cfg, mock)

	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "TestPingOK",
			wantErr: false,
		},
		{
			name:    "TestPingKO",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := app.Ping(ctx)

			if !tt.wantErr {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestGetLink(t *testing.T) {
	ctx := context.Background()
	cfg := config.NewCfg()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	shortLink := "a9tbDian"
	link := "https://practicum.yandex.ru"

	mock := mock.NewMockDatabase(ctrl)
	mock.EXPECT().GetLink(gomock.Any(), shortLink).Return(link, nil)
	mock.EXPECT().GetLink(gomock.Any(), shortLink).Return("", errTest)
	app := NewApp(ctx, cfg, mock)

	tests := []struct {
		name      string
		shortLink string
		want      string
		wantErr   bool
	}{
		{
			name:      "TestGetLinkOK",
			shortLink: shortLink,
			want:      link,
			wantErr:   false,
		},
		{
			name:      "TestGetLinkKO",
			shortLink: shortLink,
			want:      "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := app.GetLink(ctx, tt.shortLink)
			if !tt.wantErr {
				require.NoError(t, err)
				assert.Equal(t, got, tt.want, "GetLink() = %v, want %v", got, tt.want)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestCreateShortLink(t *testing.T) {
	ctx := context.Background()
	cfg := config.NewCfg()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	key := "a9tbDian"
	link := "https://practicum.yandex.ru"

	mock := mock.NewMockDatabase(ctrl)
	mock.EXPECT().SetLink(gomock.Any(), 1, key, link).Return(nil)
	mock.EXPECT().SetLink(gomock.Any(), 1, key, link).Return(errTest)

	app := NewApp(ctx, cfg, mock)

	tests := []struct {
		name    string
		link    string
		want    string
		wantErr bool
	}{
		{
			name:    "TestCreateShortLinkOK",
			link:    link,
			want:    key,
			wantErr: false,
		},
		{
			name:    "TestCreateShortLinkKO",
			link:    link,
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := app.CreateShortLink(ctx, 1, tt.link)

			if !tt.wantErr {
				require.NoError(t, err)
				assert.Equal(t, got, tt.want, "CreateShortLink() = %v, want %v", got, tt.want)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestGetShortLink(t *testing.T) {
	link := "https://practicum.yandex.ru"
	ctx := context.Background()
	cfg := config.NewCfg()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := mock.NewMockDatabase(ctrl)
	mock.EXPECT().GetShortLink(gomock.Any(), gomock.Any()).Return(gomock.Any().String(), nil)
	mock.EXPECT().GetShortLink(gomock.Any(), gomock.Any()).Return("", errTest)

	app := NewApp(ctx, cfg, mock)

	tests := []struct {
		name    string
		link    string
		want    string
		wantErr bool
	}{
		{
			name:    "TestGetShortLinkOK",
			link:    link,
			want:    gomock.Any().String(),
			wantErr: false,
		},
		{
			name:    "TestGetShortLinkKO",
			link:    link,
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := app.GetShortLink(ctx, tt.link)

			if !tt.wantErr {
				require.NoError(t, err)
				assert.Equal(t, got, tt.want, "GetShortLink() = %v, want %v", got, tt.want)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestCreateShortLinksBatch(t *testing.T) {
	reqlink := "https://practicum.yandex.ru"
	resLink := "http://localhost:8080/a9tbDian"
	ctx := context.Background()
	cfg := config.NewCfg()
	cfg.SetBaseURL("http://localhost:8080")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	reqlinksDto := make(dto.LinksDto, 0)
	reqlinksDto = append(reqlinksDto, dto.LinkDto{CorrelationID: "0", OriginalURL: reqlink})
	resLinksDto := make(dto.LinksDto, 0)
	resLinksDto = append(resLinksDto, dto.LinkDto{CorrelationID: "0", ShortURL: resLink})
	reqlinksDao := make(dao.LinksDao, 0)
	reqlinksDao = append(reqlinksDao, &dao.LinkDao{UUID: 0, CorrelationID: "0", OriginalURL: reqlink})
	resLinksDao := make(dao.LinksDao, 0)
	resLinksDao = append(resLinksDao, &dao.LinkDao{UUID: 0, CorrelationID: "0", ShortURL: "a9tbDian"})
	mock := mock.NewMockDatabase(ctrl)
	mock.EXPECT().CreateShortLinksBatch(gomock.Any(), 0, reqlinksDao).Return(resLinksDao, nil)
	mock.EXPECT().CreateShortLinksBatch(gomock.Any(), 0, reqlinksDao).Return(nil, errTest)

	app := NewApp(ctx, cfg, mock)

	tests := []struct {
		name    string
		want    dto.LinksDto
		wantErr bool
	}{
		{
			name:    "TestCreateShortLinksBatchOK",
			want:    resLinksDto,
			wantErr: false,
		},
		{
			name:    "TestCreateShortLinksBatchKO",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := app.CreateShortLinksBatch(ctx, 0, reqlinksDto)

			if !tt.wantErr {
				require.NoError(t, err)
				assert.Equal(t, got, tt.want, "CreateShortLinksBatch() = %v, want %v", got, tt.want)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestAddUser(t *testing.T) {
	ctx := context.Background()
	cfg := config.NewCfg()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := mock.NewMockDatabase(ctrl)
	mock.EXPECT().AddUser(gomock.Any()).Return(1, nil)
	mock.EXPECT().AddUser(gomock.Any()).Return(0, errTest)

	app := NewApp(ctx, cfg, mock)

	tests := []struct {
		name    string
		want    int
		wantErr bool
	}{
		{
			name:    "TestAddUserOK",
			want:    1,
			wantErr: false,
		},
		{
			name:    "TestAddUserKO",
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := app.AddUser(ctx)

			if !tt.wantErr {
				require.NoError(t, err)
				assert.Equal(t, got, tt.want, "AddUser() = %v, want %v", got, tt.want)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestGetUserURLs(t *testing.T) {
	originalURL := "https://practicum.yandex.ru"
	shortening := "a9tbDian"
	shortURL := "http://localhost:8080/a9tbDian"
	ctx := context.Background()
	cfg := config.NewCfg()
	cfg.SetBaseURL("http://localhost:8080")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	linksDao := make(dao.LinksDaoBatch, 0)
	linksDao = append(linksDao, &dao.LinkDaoBatch{ShortURL: shortening, OriginalURL: originalURL})
	linksDto := make(dto.LinksDto, 0)
	linksDto = append(linksDto, dto.LinkDto{ShortURL: shortURL, OriginalURL: originalURL})
	mock := mock.NewMockDatabase(ctrl)
	mock.EXPECT().GetUserURLs(gomock.Any(), 1).Return(linksDao, nil)
	mock.EXPECT().GetUserURLs(gomock.Any(), 1).Return(nil, errTest)

	app := NewApp(ctx, cfg, mock)

	tests := []struct {
		name    string
		want    dto.LinksDto
		wantErr bool
	}{
		{
			name:    "TestGetUserURLsOK",
			want:    linksDto,
			wantErr: false,
		},
		{
			name:    "TestGetUserURLsKO",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := app.GetUserURLs(ctx, 1)

			if !tt.wantErr {
				require.NoError(t, err)
				assert.Equal(t, got, tt.want, "GetUserURLs() = %v, want %v", got, tt.want)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestAppendChan(t *testing.T) {
	ctx := context.Background()
	cfg := config.NewCfg()
	cfg.SetTickerSeconds(1)
	cfg.SetBufferSize(1)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	slice := []string{"6YGS4ZUF", "rntA0dVu", "Yi66Tx-0", "s7L9TR66"}
	mock := mock.NewMockDatabase(ctrl)
	app := NewApp(ctx, cfg, mock)

	tests := []struct {
		name    string
		want    dto.Slice
		wantErr bool
	}{
		{
			name:    "TestAppendChanOK",
			want:    slice,
			wantErr: false,
		},
		{
			name:    "TestAppendChanKO",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app.AppendChan(ctx, 1, slice)
		})
	}
}
