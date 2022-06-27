//go:build !mobile
// +build !mobile

package zcncore

// RegisterToMiners can be used to register the wallet.
func RegisterToMiners(wallet *zcncrypto.Wallet, statusCb WalletCallback) error {
	result := make(chan *util.PostResponse)
	defer close(result)
	for _, miner := range _config.chain.Miners {
		go func(minerurl string) {
			url := minerurl + REGISTER_CLIENT
			Logger.Info(url)
			regData := map[string]string{
				"id":         wallet.ClientID,
				"public_key": wallet.ClientKey,
			}
			req, err := util.NewHTTPPostRequest(url, regData)
			if err != nil {
				Logger.Error(minerurl, "new post request failed. ", err.Error())
				return
			}
			res, err := req.Post()
			if err != nil {
				Logger.Error(minerurl, "send error. ", err.Error())
			}
			result <- res
		}(miner)
	}
	consensus := float32(0)
	for range _config.chain.Miners {
		rsp := <-result
		Logger.Debug(rsp.Url, rsp.Status)

		if rsp.StatusCode == http.StatusOK {
			consensus++
		} else {
			Logger.Debug(rsp.Body)
		}

	}
	rate := consensus * 100 / float32(len(_config.chain.Miners))
	if rate < consensusThresh {
		statusCb.OnWalletCreateComplete(StatusError, "", "rate is less than consensus")
		return fmt.Errorf("Register consensus not met. Consensus: %f, Expected: %f", rate, consensusThresh)
	}
	w, err := wallet.Marshal()
	if err != nil {
		statusCb.OnWalletCreateComplete(StatusError, w, err.Error())
		return errors.Wrap(err, "wallet encoding failed")
	}
	statusCb.OnWalletCreateComplete(StatusSuccess, w, "")
	return nil
}

