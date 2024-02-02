package storage

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

	value := "6YGS4ZUF"
	m.EXPECT().GetLink("https://ya.ru").Return(value, nil)

	val, err := m.GetLink("https://ya.ru")

	require.NoError(t, err)
	require.Equal(t, val, value)
}
