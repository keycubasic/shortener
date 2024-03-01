package postgre

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/keycubasic/shortener/internal/storage/mock"
	"github.com/stretchr/testify/require"
)

func TestGetLink(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mock.NewMockDB(ctrl)

	value := "https://ya.ru"
	m.EXPECT().GetLink("6YGS4ZUF").Return(value, nil)

	val, err := m.GetLink("6YGS4ZUF")

	require.NoError(t, err)
	require.Equal(t, val, value)
}
