package wasm

import (
	"encoding/json"
	"fmt"
	"syscall/js"
	"testing"

	"github.com/stretchr/testify/assert"

	"net/http"
	"net/http/httptest"

	"github.com/0chain/gosdk/zcncore"
)

var validMnemonic = "inside february piece turkey offer merry select combine tissue wave wet shift room afraid december gown mean brick speak grant gain become toy clown"

var walletConfig = "{\"client_id\":\"9a566aa4f8e8c342fed97c8928040a21f21b8f574e5782c28568635ba9c75a85\",\"client_key\":\"40cd10039913ceabacf05a7c60e1ad69bb2964987bc50f77495e514dc451f907c3d8ebcdab20eedde9c8f39b9a1d66609a637352f318552fb69d4b3672516d1a\",\"keys\":[{\"public_key\":\"40cd10039913ceabacf05a7c60e1ad69bb2964987bc50f77495e514dc451f907c3d8ebcdab20eedde9c8f39b9a1d66609a637352f318552fb69d4b3672516d1a\",\"private_key\":\"a3a88aad5d89cec28c6e37c2925560ce160ac14d2cdcf4a4654b2bb358fe7514\"}],\"mnemonics\":\"inside february piece turkey offer merry select combine tissue wave wet shift room afraid december gown mean brick speak grant gain become toy clown\",\"version\":\"1.0\",\"date_created\":\"2021-05-21 17:32:29.484657 +0545 +0545 m=+0.072791323\"}"

var server *httptest.Server
var miner *httptest.Server
var sharder *httptest.Server

func setup(t *testing.T) {
	// setup wallet
	TestSetWalletInfo(t)

	// setup sharder servers
	sharder = httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
			},
		),
	)

	//setup miner servers
	miner = httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
			},
		),
	)

	server = httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				n := zcncore.Network{Miners: []string{miner.URL + "/miner01"}, Sharders: []string{sharder.URL + "/sharder01"}}
				blob, err := json.Marshal(n)
				if err != nil {
					t.Fatal(err)
				}

				if _, err := w.Write(blob); err != nil {
					t.Fatal(err)
				}
			},
		),
	)

	if err := zcncore.InitZCNSDK(server.URL, "bls0chain"); err != nil {
		t.Fatal(err)
	}
}

func TestSetWalletInfo(t *testing.T) {
	setWalletInfo := js.FuncOf(SetWalletInfo)
	defer setWalletInfo.Release()

	wi := setWalletInfo.Invoke(walletConfig, "true")

	if got := wi.IsNull(); !got {
		t.Errorf("Error: %#v", wi.Get("error").String())
	}
}

func TestNetwork(t *testing.T) {
	setup(t)
	defer server.Close()

	getNetworkJSON := js.FuncOf(GetNetworkJSON)
	defer getNetworkJSON.Release()

	if got := getNetworkJSON.Invoke().String(); got != fmt.Sprintf("{\"miners\":[%#v],\"sharders\":[%#v]}", miner.URL+"/miner01", sharder.URL+"/sharder01") {
		t.Errorf("got %#v, want %#v", got, fmt.Sprintf("{\"miners\":[%#v],\"sharders\":[%#v]}", miner.URL+"/miner01", sharder.URL+"/sharder01"))
	}

	setNetwork := js.FuncOf(SetNetwork)
	defer setNetwork.Release()

	setNetwork.Invoke(miner.URL+"/miner03", sharder.URL+"/sharder03")

	if got := getNetworkJSON.Invoke().String(); got != fmt.Sprintf("{\"miners\":[%#v],\"sharders\":[%#v]}", miner.URL+"/miner03", sharder.URL+"/sharder03") {
		t.Errorf("got %#v, want %#v", got, fmt.Sprintf("{\"miners\":[%#v],\"sharders\":[%#v]}", miner.URL+"/miner03", sharder.URL+"/sharder03"))
	}
}

func TestGetMinShardersVerify(t *testing.T) {
	TestSetWalletInfo(t)
	setup(t)
	defer server.Close()
	getMinSharders := js.FuncOf(GetMinShardersVerify)
	defer getMinSharders.Release()

	if got := getMinSharders.Invoke().Int(); got != 1 {
		t.Errorf("got %#v, want %#v", got, 1)
	}
}

func TestMnemonic(t *testing.T) {
	isMnemonicValid := js.FuncOf(IsMnemonicValid)
	defer isMnemonicValid.Release()

	if got := isMnemonicValid.Invoke(validMnemonic).Bool(); !got {
		t.Errorf("got %#v, want %#v", got, true)
	}
}

func TestGetVersion(t *testing.T) {
	getVersion := js.FuncOf(GetVersion)
	defer getVersion.Release()

	if got := getVersion.Invoke().String(); got != "v1.3.0" {
		t.Errorf("got %#v, want %#v", got, "v1.3.0")
	}
}

func TestSetAuthUrl(t *testing.T) {
	setup(t)
	defer server.Close()

	TestSetWalletInfo(t)

	setAuthUrl := js.FuncOf(SetAuthUrl)
	defer setAuthUrl.Release()

	au := setAuthUrl.Invoke("miner/miner")

	if got := au.IsNull(); !got {
		t.Errorf("Error: %#v", au.Get("error").String())
	}
}

func TestSplitKeys(t *testing.T) {
	setup(t)
	defer server.Close()

	setWalletInfo := js.FuncOf(SetWalletInfo)
	defer setWalletInfo.Release()

	wi := setWalletInfo.Invoke(walletConfig, "true")

	if got := wi.IsNull(); !got {
		t.Errorf("Error: %#v", wi.Get("error").String())
	}

	splitKeys := js.FuncOf(SplitKeys)
	defer splitKeys.Release()

	au := splitKeys.Invoke("a3a88aad5d89cec28c6e37c2925560ce160ac14d2cdcf4a4654b2bb358fe7514", "2")

	if got := au.String(); got == "" {
		t.Errorf("Error: %#v", au.Get("error").String())
	}
}

func TestConversion(t *testing.T) {
	token := "100"
	ctv := js.FuncOf(ConvertToValue)
	defer ctv.Release()

	if got := ctv.Invoke(token).Int(); got != 1000000000000 {
		t.Errorf("got %#v, want %#v", got, 1000000000000)
	}

	val := ctv.Invoke(token).Int()

	ctt := js.FuncOf(ConvertToToken)
	defer ctt.Release()

	if got := ctt.Invoke(fmt.Sprintf("%d", val)).Float(); got != 100 {
		t.Errorf("got %#v, want %#v", got, 100)
	}
}

func TestEncryption(t *testing.T) {
	key := "0123456789abcdef" // must be of 16 bytes for this example to work
	var message string = "Lorem ipsum dolor sit amet"

	enc := js.FuncOf(Encrypt)
	defer enc.Release()

	emsg := enc.Invoke(key, message)

	if got := emsg.String(); got == "" {
		t.Errorf("Error: %#v", emsg.Get("error").String())
	}

	dec := js.FuncOf(Decrypt)
	defer dec.Release()

	dmsg := dec.Invoke(key, emsg.String())

	if got := dmsg.String(); got == "" {
		t.Errorf("Error: %#v", dmsg.Get("error").String())
	}

	assert.Equal(t, message, dmsg.String(), "The two message should be the same.")
}
