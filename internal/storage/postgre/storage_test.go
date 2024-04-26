package postgre

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/keycubasic/shortener/internal/app"
	"github.com/keycubasic/shortener/internal/storage/mock"
	"github.com/stretchr/testify/require"
)

func TestGetLink(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mock.NewMockDatabase(ctrl)

	value := "https://ya.ru"
	m.EXPECT().GetLink(gomock.Any(), "6YGS4ZUF").Return(value, nil)

	val, err := m.GetLink(context.TODO(), "6YGS4ZUF")

	require.NoError(t, err)
	require.Equal(t, val, value)
}

func TestSetLink(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLinks := mock.NewMockDatabase(ctrl)

	type args struct {
		ctx    context.Context
		userID int
		key    string
		link   string
	}
	tests := []struct {
		name    string
		ls      app.Database
		args    args
		ret     error
		wantErr bool
	}{
		{
			name: "TestSetLinkOK",
			ls:   mockLinks,
			args: args{
				ctx:    context.TODO(),
				userID: 1,
				key:    "OIuoiu",
				link:   "http://ya.ru",
			},
			ret:     nil,
			wantErr: false,
		},
		{
			name: "TestSetLinkKO",
			ls:   mockLinks,
			args: args{
				ctx:    context.TODO(),
				userID: 1,
				key:    "OIuoiu",
				link:   "http://ya.ru",
			},
			ret:     errors.New("oops!"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLinks.EXPECT().SetLink(tt.args.ctx, tt.args.userID, tt.args.key, tt.args.link).Return(tt.ret)
			if err := tt.ls.SetLink(tt.args.ctx, tt.args.userID, tt.args.key, tt.args.link); (err != nil) != tt.wantErr {
				t.Errorf("Links.SetLink() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
