package cmd

import (
	"fmt"
	"os"
	"sync"

	"github.com/0chain/gosdk/zcncore"
	"github.com/spf13/cobra"
)

var recoverwalletcmd = &cobra.Command{
	Use:   "recoverwallet",
	Short: "Recover wallet",
	Long:  `Recover wallet from the previously stored mnemonic`,
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		fflags := cmd.Flags()
		if fflags.Changed("mnemonic") == false {
			fmt.Println("\nError: Mnemonic not provided\n")
			return
		}
		mnemonic := cmd.Flag("mnemonic").Value.String()
		if zcncore.IsMnemonicValid(mnemonic) == false {
			fmt.Println("\nError: Invalid mnemonic\n")
			return
		}
		wg := &sync.WaitGroup{}
		statusBar := &ZCNStatus{wg: wg}
		wg.Add(1)
		err := zcncore.RecoverWallet(mnemonic, numKeys, statusBar)
		if err == nil {
			wg.Wait()
		} else {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		if len(statusBar.walletString) == 0 || !statusBar.success {
			fmt.Println("Error recovering the wallet." + statusBar.errMsg)
			os.Exit(1)
		}
		var walletFilePath string
		if &walletFile != nil && len(walletFile) > 0 {
			walletFilePath = getConfigDir() + "/" + walletFile
		} else {
			walletFilePath = getConfigDir() + "/wallet.txt"
		}
		clientConfig = string(statusBar.walletString)
		file, err := os.Create(walletFilePath)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		defer file.Close()
		fmt.Fprintf(file, clientConfig)
		fmt.Println("\nWallet recovered!!\n")
		return
	},
}

var getbalancecmd = &cobra.Command{
	Use:   "getbalance",
	Short: "get balance from sharders",
	Long:  `get balance from sharders`,
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		wg := &sync.WaitGroup{}
		statusBar := &ZCNStatus{wg: wg}
		wg.Add(1)
		err := zcncore.GetBalance(statusBar)
		if err == nil {
			wg.Wait()
		} else {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		if statusBar.success {
			fmt.Println("\nBalance:", zcncore.ConvertToToken(statusBar.balance), "\n")
		} else {
			fmt.Println("\nGet balance failed. " + statusBar.errMsg + "\n")
		}
		return
	},
}

var sendcmd = &cobra.Command{
	Use:   "send",
	Short: "Send ZCN token to another wallet",
	Long: `Send ZCN token to another wallet.
	        <toclientID> <token> <description>`,
	Args: cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		fflags := cmd.Flags()
		if fflags.Changed("toclientID") == false {
			fmt.Println("Error: toclientID flag is missing")
			return
		}
		if fflags.Changed("token") == false {
			fmt.Println("Error: token flag is missing")
			return
		}
		if fflags.Changed("desc") == false {
			fmt.Println("Error: Description flag is missing")
			return
		}
		toclientID := cmd.Flag("toclientID").Value.String()
		token, err := cmd.Flags().GetFloat64("token")
		desc := cmd.Flag("desc").Value.String()
		wg := &sync.WaitGroup{}
		statusBar := &ZCNStatus{wg: wg}
		txn, err := zcncore.NewTransaction(statusBar)
		if err != nil {
			fmt.Println(err)
			return
		}
		wg.Add(1)
		err = txn.Send(toclientID, zcncore.ConvertToValue(token), desc)
		if err == nil {
			wg.Wait()
		} else {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		if statusBar.success {
			statusBar.success = false
			wg.Add(1)
			err := txn.Verify()
			if err == nil {
				wg.Wait()
			} else {
				fmt.Println(err.Error())
				os.Exit(1)
			}
			if statusBar.success {
				fmt.Println("\nSend token success\n")
				return
			}
		}
		fmt.Println("Send token failed. " + statusBar.errMsg)
		return
	},
}

var faucetcmd = &cobra.Command{
	Use:   "faucet",
	Short: "Faucet smart contract",
	Long: `Faucet smart contract.
	        <methodName> <input>`,
	Args: cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		fflags := cmd.Flags()
		if fflags.Changed("methodName") == false {
			fmt.Println("Error: Methodname flag is missing")
			return
		}
		if fflags.Changed("input") == false {
			fmt.Println("Error: Input flag is missing")
			return
		}

		methodName := cmd.Flag("methodName").Value.String()
		input := cmd.Flag("input").Value.String()
		wg := &sync.WaitGroup{}
		statusBar := &ZCNStatus{wg: wg}
		txn, err := zcncore.NewTransaction(statusBar)
		if err != nil {
			fmt.Println(err)
			return
		}
		wg.Add(1)
		err = txn.ExecuteFaucetSC(methodName, []byte(input))
		if err == nil {
			wg.Wait()
		} else {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		if statusBar.success {
			statusBar.success = false
			wg.Add(1)
			err := txn.Verify()
			if err == nil {
				wg.Wait()
			} else {
				fmt.Println(err.Error())
				os.Exit(1)
			}
			if statusBar.success {
				fmt.Println("\nExecute faucet smart contract success\n")
				return
			}
		}
		fmt.Println("\nExecute faucet smart contract failed. " + statusBar.errMsg + "\n")
		return
	},
}

var lockcmd = &cobra.Command{
	Use:   "lock",
	Short: "Lock tokens",
	Long: `Lock tokens .
	        <tokens> <[durationHr] [durationMin]>`,
	Args: cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		fflags := cmd.Flags()
		if fflags.Changed("token") == false {
			fmt.Println("Error: token flag is missing")
			return
		}
		if (fflags.Changed("durationHr") == false) &&
			(fflags.Changed("durationMin") == false) {
			fmt.Println("Error: durationHr and durationMin flag is missing. atleast one is required")
			return
		}
		token, err := cmd.Flags().GetFloat64("token")
		if err != nil {
			fmt.Println("Error: invalid number of tokens")
			return
		}
		durationHr := int64(0)
		durationHr, err = cmd.Flags().GetInt64("durationHr")
		durationMin := int(0)
		durationMin, err = cmd.Flags().GetInt("durationMin")
		if (durationHr < 1) && (durationMin < 1) {
			fmt.Println("Error: invalid duration")
		}
		wg := &sync.WaitGroup{}
		statusBar := &ZCNStatus{wg: wg}
		txn, err := zcncore.NewTransaction(statusBar)
		if err != nil {
			fmt.Println(err)
			return
		}
		wg.Add(1)
		err = txn.LockTokens(zcncore.ConvertToValue(token), durationHr, durationMin)
		if err == nil {
			wg.Wait()
		} else {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		if statusBar.success {
			statusBar.success = false
			wg.Add(1)
			err := txn.Verify()
			if err == nil {
				wg.Wait()
			} else {
				fmt.Println(err.Error())
				os.Exit(1)
			}
			if statusBar.success {
				fmt.Printf("\nTokens (%f) locked successfully\n", token)
				return
			}
		}
		fmt.Println("\nFailed to lock tokens. " + statusBar.errMsg + "\n")
		return
	},
}

var unlockcmd = &cobra.Command{
	Use:   "unlock",
	Short: "Unlock tokens",
	Long: `Unlock previously locked tokens .
	        <poolid>`,
	Args: cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		fflags := cmd.Flags()
		if fflags.Changed("poolid") == false {
			fmt.Println("Error: poolid flag is missing")
			return
		}
		poolid := cmd.Flag("poolid").Value.String()
		wg := &sync.WaitGroup{}
		statusBar := &ZCNStatus{wg: wg}
		txn, err := zcncore.NewTransaction(statusBar)
		if err != nil {
			fmt.Println(err)
			return
		}
		wg.Add(1)
		err = txn.UnlockTokens(poolid)
		if err == nil {
			wg.Wait()
		} else {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		if statusBar.success {
			statusBar.success = false
			wg.Add(1)
			err := txn.Verify()
			if err == nil {
				wg.Wait()
			} else {
				fmt.Println(err.Error())
				os.Exit(1)
			}
			if statusBar.success {
				fmt.Printf("\nUnlock token success\n")
				return
			}
		}
		fmt.Println("\nFailed to unlock tokens. " + statusBar.errMsg + "\n")
		return
	},
}

var lockconfigcmd = &cobra.Command{
	Use:   "lockconfig",
	Short: "Get lock configuration",
	Long:  `Get lock configuration`,
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		wg := &sync.WaitGroup{}
		statusBar := &ZCNStatus{wg: wg}
		wg.Add(1)
		err := zcncore.GetLockConfig(statusBar)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		wg.Wait()
		if statusBar.success {
			fmt.Printf("\nConfiguration:\n %v\n", statusBar.errMsg)
			return
		}
		fmt.Println("\nFailed to get lock config." + statusBar.errMsg + "\n")
		return
	},
}

var getlockedtokenscmd = &cobra.Command{
	Use:   "getlockedtokens",
	Short: "Get locked tokens",
	Long:  `Get locked tokens`,
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		wg := &sync.WaitGroup{}
		statusBar := &ZCNStatus{wg: wg}
		wg.Add(1)
		err := zcncore.GetLockedTokens(statusBar)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		wg.Wait()
		if statusBar.success {
			fmt.Printf("\nLocked tokens:\n %v\n", statusBar.errMsg)
			return
		}
		fmt.Println("\nFailed to get locked tokens." + statusBar.errMsg + "\n")
		return
	},
}

func init() {
	rootCmd.AddCommand(recoverwalletcmd)
	rootCmd.AddCommand(getbalancecmd)
	rootCmd.AddCommand(sendcmd)
	rootCmd.AddCommand(faucetcmd)
	rootCmd.AddCommand(lockcmd)
	rootCmd.AddCommand(unlockcmd)
	rootCmd.AddCommand(lockconfigcmd)
	rootCmd.AddCommand(getlockedtokenscmd)
	recoverwalletcmd.PersistentFlags().String("mnemonic", "", "mnemonic")
	recoverwalletcmd.MarkFlagRequired("mnemonic")
	sendcmd.PersistentFlags().String("toclientID", "", "toclientID")
	sendcmd.PersistentFlags().Float64("token", 0, "Token to send")
	sendcmd.PersistentFlags().String("desc", "", "Description")
	sendcmd.MarkFlagRequired("toclientID")
	sendcmd.MarkFlagRequired("token")
	sendcmd.MarkFlagRequired("desc")
	faucetcmd.PersistentFlags().String("methodName", "", "methodName")
	faucetcmd.PersistentFlags().String("input", "", "input")
	faucetcmd.MarkFlagRequired("methodName")
	faucetcmd.MarkFlagRequired("input")
	lockcmd.PersistentFlags().Float64("token", 0, "Number to tokens to lock")
	lockcmd.PersistentFlags().Int64("durationHr", 0, "Duration Hours to lock")
	lockcmd.PersistentFlags().Int("durationMin", 0, "Duration Mins to lock")
	lockcmd.MarkFlagRequired("token")
	unlockcmd.PersistentFlags().String("poolid", "", "Poolid - hash of the locked transaction")
	unlockcmd.MarkFlagRequired("poolid")
}
