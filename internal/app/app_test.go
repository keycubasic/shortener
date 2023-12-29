package app

import (
	"testing"

	"github.com/keycubasic/shortener/internal/storage"
)

func TestCreateShortLink(t *testing.T) {
	type args struct {
		body []byte
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "TestCreateShortLinkOK",
			args:    args{[]byte("https://practicum.yandex.ru")},
			want:    "a9tbDian",
			wantErr: false,
		},
	}

	db, _ := storage.NewLinks()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateShortLink(db, tt.args.body)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateShortLink() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CreateShortLink() = %v, want %v", got, tt.want)
			}
		})
	}
}
