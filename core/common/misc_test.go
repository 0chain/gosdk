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

func TestFormatStatic(t *testing.T) {
	amount := int64(129382129321)
	require.Equal(t, "12.938 ZCN", FormatStatic(amount, "ZCN"))
	require.Equal(t, "12938.213 mZCN", FormatStatic(amount, "mZCN"))
	require.Equal(t, "12938212.932 uZCN", FormatStatic(amount, "uZCN"))
	require.Equal(t, "129382129321 SAS", FormatStatic(amount, "SAS"))
}

func TestAutoFormat(t *testing.T) {
	require.Equal(t, "239 SAS", AutoFormatStatic(239))
	require.Equal(t, "2.736 uZCN", AutoFormatStatic(27361))
	require.Equal(t, "2.387 mZCN", AutoFormatStatic(23872013))
	require.Equal(t, "20.383 ZCN", AutoFormatStatic(203827162834))
}

func TestAutoFormatStatic(t *testing.T) {
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
