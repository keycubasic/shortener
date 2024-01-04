package app

import (
	"testing"

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

	db := ConnectStorage()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateShortLink(db, tt.args.link)
			require.NoError(t, err)
			if tt.wantErr {
				t.Errorf("CreateShortLink() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, got, tt.want, "CreateShortLink() = %v, want %v", got, tt.want)
		})
	}
}
