package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToTransactionState(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected TransactionState
		wantErr  bool
	}{
		{"win", "win", TransactionStateWin, false},
		{"lose", "lose", TransactionStateLose, false},
		{"invalid", "draw", "", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			state, err := ToTransactionState(tc.input)

			if tc.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expected, state)
		})
	}
}

func TestToSourceType(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected SourceType
		wantErr  bool
	}{
		{"game", "game", SourceTypeGame, false},
		{"server", "server", SourceTypeServer, false},
		{"payment", "payment", SourceTypePayment, false},
		{"invalid", "unknown", "", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			source, err := ToSourceType(tc.input)

			if tc.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expected, source)
		})
	}
}
