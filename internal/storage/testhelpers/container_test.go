package testhelpers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreatePostgresContainer(t *testing.T) {
	_, err := CreatePostgresContainer(context.TODO())
	require.NoError(t, err)
}
