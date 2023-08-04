package transaction

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/0chain/common/core/encryption"
	"github.com/0chain/errors"
	"github.com/0chain/gosdk/core/conf"
	"github.com/0chain/gosdk/core/resty"
	"github.com/0chain/gosdk/core/util"
)

const retriesCount = 30

type OptimisticVerifier struct {
	allSharders []string
	sharders    []string
	options     []resty.Option
}

func NewOptimisticVerifier(sharders []string) *OptimisticVerifier {
	//initialize resty
	header := map[string]string{
		"Content-Type":                "application/json; charset=utf-8",
		"Access-Control-Allow-Origin": "*",
	}

	transport := createTransport(resty.DefaultDialTimeout)

	options := []resty.Option{
		resty.WithRetry(resty.DefaultRetry),
		resty.WithHeader(header),
		resty.WithTransport(transport),
	}

	return &OptimisticVerifier{
		allSharders: sharders,
		options:     options,
	}
}

func (v *OptimisticVerifier) VerifyTransactionOptimistic(txnHash string) (*Transaction, error) {
	cfg, err := conf.GetClientConfig()
	if err != nil {

		return nil, err
	}

	//refresh sharders
	v.sharders = v.allSharders

	//amount of sharders to query
	minNumConfirmation := int(math.Ceil(float64(cfg.MinConfirmation*len(v.sharders)) / 100))
	if minNumConfirmation > len(v.sharders) {
		return nil, errors.New("verify_optimistic", "wrong number of min_confirmations")
	}
	shuffled := util.Shuffle(v.sharders)[:minNumConfirmation]

	//prepare urls for confirmation request
	urls := make([]string, 0, len(shuffled))
	mappedSharders := make(map[string]string)
	for _, sharder := range shuffled {
		url := fmt.Sprintf("%v/%v%v", sharder, TXN_VERIFY_URL, txnHash)
		urls = append(urls, url)
		mappedSharders[url] = sharder
	}

	var url string
	var chain []*RoundBlockHeader
	var txn *Transaction
	r := resty.New(v.options...).Then(func(req *http.Request, resp *http.Response, respBody []byte, cf context.CancelFunc, err error) error {
		if err != nil { //network issue
			return err
		}

		if resp.StatusCode != 200 {
			return errors.Throw(ErrInvalidRequest, strconv.Itoa(resp.StatusCode)+": "+resp.Status)
		}

		//parse response
		var objmap map[string]json.RawMessage
		err = json.Unmarshal(respBody, &objmap)
		if err != nil {
			return err
		}
		txnRawJSON, ok := objmap["txn"]
		// txn data is found, success
		if !ok {
			return errors.New("handle_response", "bad transaction response")
		}
		merklePathRawJSON, ok := objmap["merkle_tree_path"]
		if !ok {
			return errors.New("handle_response", "bad merkle_tree_path response")
		}

		txn = &Transaction{}
		err = json.Unmarshal(txnRawJSON, txn)
		if err != nil {
			return err
		}

		b := &RoundBlockHeader{}
		err = json.Unmarshal(respBody, b)
		if err != nil {
			return err
		}
		err = validateBlockHash(b)
		if err != nil {
			return err
		}

		err = verifyMerklePath(merklePathRawJSON, txn.Hash, b.MerkleTreeRoot)
		if err != nil {
			return err
		}

		url = req.URL.String()
		chain = append(chain, b)
		return nil
	})

	retries := 0
	ticker := time.NewTicker(time.Second)
L:
	//loop query confirmation
	for retries < retriesCount {
		<-ticker.C
		retries++
		r.DoGet(context.TODO(), urls...)
		//need single valid confirmation
		errs := r.First()
		if len(errs) == 0 {
			break L
		}
	}

	if len(chain) == 0 {
		return nil, errors.Newf("verify", "can't get confirmation after %v retries", retriesCount)
	}

	//remove current sharder from the list to avoid building chain with it
	toDelete := mappedSharders[url]
	for i, s := range v.sharders {
		if s == toDelete {
			v.sharders = append(v.sharders[:i], v.sharders[i+1:]...)
			break
		}
	}

	err = v.checkConfirmation(chain)
	if err != nil {
		return nil, err
	}

	return txn, err
}

func (v *OptimisticVerifier) checkConfirmation(chain []*RoundBlockHeader) error {
	cfg, err := conf.GetClientConfig()
	if err != nil {

		return err
	}

	//build blockchain starting from confirmation block
	curRound := chain[0].Round
	rb := resty.New(v.options...).Then(func(req *http.Request, resp *http.Response, respBody []byte, cf context.CancelFunc, err error) error {
		if err != nil { //network issue
			return err
		}

		if resp.StatusCode != 200 {
			return errors.Throw(ErrInvalidRequest, strconv.Itoa(resp.StatusCode)+": "+resp.Status)
		}

		curBlock := &Block{}
		err = json.Unmarshal(respBody, &curBlock)
		if err != nil {
			return err
		}

		//get tail block and check that current extends it
		prevBlock := chain[len(chain)-1]
		if prevBlock.Hash == curBlock.PrevHash && prevBlock.Round+1 == curBlock.Round {
			blockHeader := &RoundBlockHeader{
				Version:               curBlock.Version,
				CreationDate:          curBlock.CreationDate,
				Hash:                  curBlock.Hash,
				PreviousBlockHash:     curBlock.PrevHash,
				MinerID:               curBlock.MinerID,
				Round:                 curBlock.Round,
				RoundRandomSeed:       curBlock.RoundRandomSeed,
				MerkleTreeRoot:        curBlock.MerkleTreeRoot,
				StateChangesCount:     curBlock.StateChangesCount,
				StateHash:             curBlock.StateHash,
				ReceiptMerkleTreeRoot: curBlock.ReceiptMerkleTreeRoot,
				NumberOfTxns:          int64(curBlock.NumTxns),
			}
			err = validateBlockHash(blockHeader)
			if err != nil {
				return err
			}

			chain = append(chain, blockHeader)
			return nil
		}
		return errors.New("get_block", "wrong block")
	})

	//query for blocks until ConfirmationChainLength is built or every sharder is queried
	for len(chain) < cfg.ConfirmationChainLength && len(v.sharders) > 0 {
		//for every new block create sharder list to query
		rand := util.NewRand(len(v.sharders))
		//iterate through all sharders sequentially to get next block
		for {
			next, err := rand.Next()
			if err != nil {
				return errors.New("get_round_block", "can't get round block, blockchain might be stuck")
			}

			cur := v.sharders[next]
			burl := fmt.Sprintf("%v/%v%v", cur, BLOCK_BY_ROUND_URL, curRound+1)
			rb.DoGet(context.TODO(), burl)

			wait := rb.Wait()
			if len(wait) != 0 {
				continue
			}
			//exclude sharder if it gave block, we do it to avoid building blockchain from single sharder
			v.sharders = append(v.sharders[:next], v.sharders[next+1:]...)
			curRound++
			break
		}
	}

	return nil
}

func verifyMerklePath(merklePathRawJSON json.RawMessage, txnHash string, merkleRoot string) error {
	merklePath := &util.MTPath{}
	err := json.Unmarshal(merklePathRawJSON, merklePath)
	if err != nil {
		return err
	}
	if !util.VerifyMerklePath(txnHash, merklePath, merkleRoot) {
		return errors.New("handle_response", "invalid merkle path")
	}
	return nil
}

func validateBlockHash(b *RoundBlockHeader) error {
	hashBuilder := strings.Builder{}
	hashBuilder.WriteString(b.MinerID)
	hashBuilder.WriteString(":")
	hashBuilder.WriteString(b.PreviousBlockHash)
	hashBuilder.WriteString(":")
	hashBuilder.WriteString(strconv.FormatInt(b.CreationDate, 10))
	hashBuilder.WriteString(":")
	hashBuilder.WriteString(strconv.FormatInt(b.Round, 10))
	hashBuilder.WriteString(":")
	hashBuilder.WriteString(strconv.FormatInt(b.RoundRandomSeed, 10))
	hashBuilder.WriteString(":")
	hashBuilder.WriteString(strconv.Itoa(b.StateChangesCount))
	hashBuilder.WriteString(":")
	hashBuilder.WriteString(b.MerkleTreeRoot)
	hashBuilder.WriteString(":")
	hashBuilder.WriteString(b.ReceiptMerkleTreeRoot)
	//todo handling of magic block here
	hash := encryption.Hash(hashBuilder.String())
	if hash != b.Hash {
		return errors.New("handle_response", "invalid block hash")
	}
	return nil
}

// VerifyTransaction query transaction status from sharders, and verify it by mininal confirmation
func VerifyTransaction(txnHash string, sharders []string) (*Transaction, error) {
	cfg, err := conf.GetClientConfig()
	if err != nil {
		return nil, err
	}

	if cfg.VerifyOptimistic {
		ov := NewOptimisticVerifier(sharders)
		return ov.VerifyTransactionOptimistic(txnHash)
	} else {
		return VerifyTransactionTrusted(txnHash, sharders)
	}
}

// VerifyTransaction query transaction status from sharders, and verify it by mininal confirmation
func VerifyTransactionTrusted(txnHash string, sharders []string) (*Transaction, error) {

	cfg, err := conf.GetClientConfig()
	if err != nil {

		return nil, err
	}

	numSharders := len(sharders)

	if numSharders == 0 {
		return nil, ErrNoAvailableSharder
	}

	minNumConfirmation := int(math.Ceil(float64(cfg.MinConfirmation*numSharders) / 100))

	rand := util.NewRand(numSharders)

	selectedSharders := make([]string, 0, minNumConfirmation+1)

	// random pick minNumConfirmation+1 first
	for i := 0; i <= minNumConfirmation; i++ {
		n, err := rand.Next()

		if err != nil {
			break
		}

		selectedSharders = append(selectedSharders, sharders[n])
	}

	numSuccess := 0

	var retTxn *Transaction

	//leave first item for ErrTooLessConfirmation
	var msgList = make([]string, 1, numSharders)

	urls := make([]string, 0, len(selectedSharders))

	for _, sharder := range selectedSharders {
		urls = append(urls, fmt.Sprintf("%v/%v%v", sharder, TXN_VERIFY_URL, txnHash))
	}

	header := map[string]string{
		"Content-Type":                "application/json; charset=utf-8",
		"Access-Control-Allow-Origin": "*",
	}

	transport := createTransport(resty.DefaultDialTimeout)

	options := []resty.Option{
		resty.WithRetry(resty.DefaultRetry),
		resty.WithHeader(header),
		resty.WithTransport(transport),
	}

	r := resty.New(options...).
		Then(func(req *http.Request, resp *http.Response, respBody []byte, cf context.CancelFunc, err error) error {
			url := req.URL.String()

			if err != nil { //network issue
				msgList = append(msgList, err.Error())
				return err
			}

			if resp.StatusCode != 200 {
				msgList = append(msgList, url+": ["+strconv.Itoa(resp.StatusCode)+"] "+string(respBody))
				return errors.Throw(ErrInvalidRequest, strconv.Itoa(resp.StatusCode)+": "+resp.Status)
			}

			var objmap map[string]json.RawMessage
			err = json.Unmarshal(respBody, &objmap)
			if err != nil {
				msgList = append(msgList, "json: "+string(respBody))
				return err
			}
			txnRawJSON, ok := objmap["txn"]

			// txn data is found, success
			if ok {
				txn := &Transaction{}
				err = json.Unmarshal(txnRawJSON, txn)
				if err != nil {
					msgList = append(msgList, "json: "+string(txnRawJSON))
					return err
				}
				if len(txn.Signature) > 0 {
					retTxn = txn
				}
				numSuccess++

			} else {
				// txn data is not found, but get block_hash, success
				if _, ok := objmap["block_hash"]; ok {
					numSuccess++
				} else {
					// txn and block_hash
					msgList = append(msgList, fmt.Sprintf("Sharder does not have the block summary with url: %s, contents: %s", url, string(respBody)))
				}

			}

			return nil
		})

	for {
		r.DoGet(context.TODO(), urls...)

		r.Wait()

		if numSuccess >= minNumConfirmation {
			break
		}

		// pick one more sharder to query transaction
		n, err := rand.Next()

		if errors.Is(err, util.ErrNoItem) {
			break
		}

		urls = []string{fmt.Sprintf("%v/%v%v", sharders[n], TXN_VERIFY_URL, txnHash)}

	}

	if numSuccess > 0 && numSuccess >= minNumConfirmation {
		if retTxn == nil {
			return nil, errors.Throw(ErrNoTxnDetail, strings.Join(msgList, "\r\n"))
		}
		return retTxn, nil
	}

	msgList[0] = fmt.Sprintf("min_confirmation is %v%%, but got %v/%v sharders", cfg.MinConfirmation, numSuccess, numSharders)
	return nil, errors.Throw(ErrTooLessConfirmation, strings.Join(msgList, "\r\n"))

}
