package common

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFormat(t *testing.T) {
	token := Balance(129382129321)
	require.Equal(t, "12.938 ZCN", token.Format(ZCN))
	require.Equal(t, "12938.213 mZCN", token.Format(MZCN))
	require.Equal(t, "12938212.932 uZCN", token.Format(UZCN))
	require.Equal(t, "129382129321 SAS", token.Format(SAS))
}

func TestAutoFormat(t *testing.T) {
	require.Equal(t, "239 SAS", Balance(239).AutoFormat())
	require.Equal(t, "2.736 uZCN", Balance(27361).AutoFormat())
	require.Equal(t, "2.387 mZCN", Balance(23872013).AutoFormat())
	require.Equal(t, "20.383 ZCN", Balance(203827162834).AutoFormat())
}

func TestParseBalance(t *testing.T) {
	b, err := ParseBalance("12.938 ZCN")
	require.NoError(t, err)
	require.Equal(t, Balance(12.938*1e10), b)

	b, err = ParseBalance("12.938 mzcn")
	require.NoError(t, err)
	require.Equal(t, Balance(12.938*1e7), b)

	b, err = ParseBalance("12.938 uZCN")
	require.NoError(t, err)
	require.Equal(t, Balance(12.938*1e4), b)

	b, err = ParseBalance("122389 sas")
	require.NoError(t, err)
	require.Equal(t, Balance(122389*1e0), b)

	_, err = ParseBalance("10 ")
	require.EqualError(t, err, "invalid input: 10 ")

	_, err = ParseBalance("10 zwe")
	require.EqualError(t, err, "invalid input: 10 zwe")

	_, err = ParseBalance(" 10 zcn ")
	require.EqualError(t, err, "invalid input:  10 zcn ")
}
