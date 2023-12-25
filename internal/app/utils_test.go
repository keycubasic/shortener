package app

import (
	"testing"
)

func TestGetKeyFromPath(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "GetKeyFromPathOK",
			args: args{"/a9tbDian"},
			want: "a9tbDian",
			wantErr: false,
		},
		{
			name: "GetKeyFromPathKO empty",
			args: args{""},
			want: "",
			wantErr: true,
		},
		{
			name: "GetKeyFromPathKO multiple",
			args: args{"/a9tbDian/3"},
			want: "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetKeyFromPath(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetKeyFromPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetKeyFromPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

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
			name: "TestCreateShortLinkOK",
			args: args{[]byte("https://practicum.yandex.ru")},
			want: "a9tbDian",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateShortLink(tt.args.body)
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
