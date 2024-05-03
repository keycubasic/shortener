// Package mapper contains functions for mapping data objects.
package mapper

import (
	"reflect"
	"testing"

	"github.com/keycubasic/shortener/internal/dao"
	"github.com/keycubasic/shortener/internal/dto"
)

func TestLinkDtoToLinkDao(t *testing.T) {
	type args struct {
		linkDto dto.LinkDto
	}
	tests := []struct {
		name string
		args args
		want *dao.LinkDao
	}{
		{
			name: "TestLinkDtoToLinkDaoOK",
			args: args{linkDto: *new(dto.LinkDto)},
			want: new(dao.LinkDao),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LinkDtoToLinkDao(tt.args.linkDto); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LinkDtoToLinkDao() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLinkDaoToLinkDto(t *testing.T) {
	type args struct {
		linkDao dao.LinkDao
		baseURL string
	}
	tests := []struct {
		name string
		args args
		want *dto.LinkDto
	}{
		{
			name: "TestLinkDaoToLinkDtoOK",
			args: args{linkDao: *new(dao.LinkDao), baseURL: ""},
			want: &dto.LinkDto{ShortURL: "/"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LinkDaoToLinkDto(tt.args.linkDao, tt.args.baseURL); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LinkDaoToLinkDto() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLinkDaoToLinkDtoURLs(t *testing.T) {
	type args struct {
		linkDao dao.LinkDaoBatch
		baseURL string
	}
	tests := []struct {
		name string
		args args
		want *dto.LinkDto
	}{
		{
			name: "TestLinkDaoToLinkDtoURLsOK",
			args: args{linkDao: *new(dao.LinkDaoBatch), baseURL: ""},
			want: &dto.LinkDto{ShortURL: "/"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LinkDaoToLinkDtoURLs(tt.args.linkDao, tt.args.baseURL); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LinkDaoToLinkDtoURLs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLinkToLinkDao(t *testing.T) {
	type args struct {
		userID int
		link   string
	}
	tests := []struct {
		name string
		args args
		want *dao.LinkDao
	}{
		{
			name: "TestLinkToLinkDaoOK",
			args: args{userID: 0, link: ""},
			want: &dao.LinkDao{DeletedFlag: true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LinkToLinkDao(tt.args.userID, tt.args.link); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LinkToLinkDao() = %v, want %v", got, tt.want)
			}
		})
	}
}
