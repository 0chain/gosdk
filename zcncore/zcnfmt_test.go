package zcncore

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFormat(t *testing.T) {
	token := Balance(129382129321)
	require.Equal(t, "129.382 ZCN", token.Format(ZCN))
	require.Equal(t, "129382.129 mZCN", token.Format(MZCN))
	require.Equal(t, "129382129.321 uZCN", token.Format(UZCN))
	require.Equal(t, "129382129321 SAS", token.Format(SAS))
}

func TestAutoFormat(t *testing.T) {
	require.Equal(t, "239 SAS", Balance(239).AutoFormat())
	require.Equal(t, "27.361 uZCN", Balance(27361).AutoFormat())
	require.Equal(t, "23.872 mZCN", Balance(23872013).AutoFormat())
	require.Equal(t, "203.827 ZCN", Balance(203827162834).AutoFormat())
}
