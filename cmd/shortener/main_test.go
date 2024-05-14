// Main package of application
package main

import (
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestReadBuildInfo(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "TestReadBuildInfoOK",
			args: args{key: "empty"},
			want: "N/A",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := readBuildInfo(tt.args.key); got != tt.want {
				t.Errorf("readBuildInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}
