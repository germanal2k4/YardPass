package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeResidentPhone(t *testing.T) {
	t.Run("nil_and_empty", func(t *testing.T) {
		got, err := NormalizeResidentPhone(nil)
		require.NoError(t, err)
		assert.Nil(t, got)
		s := ""
		got, err = NormalizeResidentPhone(&s)
		require.NoError(t, err)
		assert.Nil(t, got)
	})
	t.Run("formats", func(t *testing.T) {
		cases := []struct {
			in   string
			want string
		}{
			{"+7 (900) 123-45-67", "+79001234567"},
			{"89001234567", "+79001234567"},
			{"9001234567", "+79001234567"},
		}
		for _, tc := range cases {
			in := tc.in
			got, err := NormalizeResidentPhone(&in)
			require.NoError(t, err, tc.in)
			require.NotNil(t, got)
			assert.Equal(t, tc.want, *got)
		}
	})
	t.Run("invalid", func(t *testing.T) {
		bad := "123"
		_, err := NormalizeResidentPhone(&bad)
		require.Error(t, err)
	})
}
