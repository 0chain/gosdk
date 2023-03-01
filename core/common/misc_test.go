package common

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFormat(t *testing.T) {
	token := Balance(129382129321)
	formattedToken, err := token.Format(ZCN)
	require.Equal(t, "12.938 ZCN", formattedToken)
	require.NoError(t, err)

	formattedToken, err = token.Format(MZCN)
	require.Equal(t, "12938.213 mZCN", formattedToken)
	require.NoError(t, err)

	formattedToken, err = token.Format(UZCN)
	require.Equal(t, "12938212.932 uZCN", formattedToken)
	require.NoError(t, err)

	formattedToken, err = token.Format(SAS)
	require.Equal(t, "129382129321 SAS", formattedToken)
	require.NoError(t, err)

	_, err = token.Format(5)
	require.EqualError(t, err, "undefined balance unit: 5")
}

func TestFormatStatic(t *testing.T) {
	amount := int64(129382129321)
	zcnAmount, err := FormatStatic(amount, "ZCN")
	require.Equal(t, "12.938 ZCN", zcnAmount)
	require.NoError(t, err)

	mZCNAmount, err := FormatStatic(amount, "mZCN")
	require.Equal(t, "12938.213 mZCN", mZCNAmount)
	require.NoError(t, err)

	uZCN, err := FormatStatic(amount, "uZCN")
	require.Equal(t, "12938212.932 uZCN", uZCN)
	require.NoError(t, err)

	sas, err := FormatStatic(amount, "SAS")
	require.Equal(t, "129382129321 SAS", sas)
	require.NoError(t, err)
}

func TestAutoFormat(t *testing.T) {
	autoFormatValue, err := AutoFormatStatic(239)
	require.Equal(t, "239 SAS", autoFormatValue)
	require.NoError(t, err)

	autoFormatValue, err = AutoFormatStatic(27361)
	require.Equal(t, "2.736 uZCN", autoFormatValue)
	require.NoError(t, err)

	autoFormatValue, err = AutoFormatStatic(23872013)
	require.Equal(t, "2.387 mZCN", autoFormatValue)
	require.NoError(t, err)

	autoFormatValue, err = AutoFormatStatic(203827162834)
	require.Equal(t, "20.383 ZCN", autoFormatValue)
	require.NoError(t, err)
}

func TestAutoFormatStatic(t *testing.T) {
	autoFormatValue, err := Balance(239).AutoFormat()
	require.Equal(t, "239 SAS", autoFormatValue)
	require.NoError(t, err)

	autoFormatValue, err = Balance(27361).AutoFormat()
	require.Equal(t, "2.736 uZCN", autoFormatValue)
	require.NoError(t, err)

	autoFormatValue, err = Balance(23872013).AutoFormat()
	require.Equal(t, "2.387 mZCN", autoFormatValue)
	require.NoError(t, err)

	autoFormatValue, err = Balance(203827162834).AutoFormat()
	require.Equal(t, "20.383 ZCN", autoFormatValue)
	require.NoError(t, err)
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

func TestToBalance(t *testing.T) {
	expectedBalance := Balance(20100000000)
	balance, err := ToBalance(2.01)
	require.Equal(t, expectedBalance, balance)
	require.NoError(t, err)

	expectedBalance = Balance(30099999999)
	balance, err = ToBalance(3.0099999999)
	require.Equal(t, expectedBalance, balance)
	require.NoError(t, err)
}

func TestToToken(t *testing.T) {
	b := Balance(12.938 * 1e12)
	token, err := b.ToToken()
	require.Equal(t, 1293.8, token)
	require.NoError(t, err)

	b = Balance(12.938 * 1e5)
	token, err = b.ToToken()
	require.Equal(t, 0.00012938, token)
	require.NoError(t, err)
}
