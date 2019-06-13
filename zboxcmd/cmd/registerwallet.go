package cmd

import (
	"fmt"
	"sync"

	"0chain.net/clientsdk/zcncore"
	"github.com/spf13/cobra"
)

// registerWalletCmd represents the register wallet command
var registerWalletCmd = &cobra.Command{
	Use:   "register",
	Short: "Registers the wallet with the blockchain",
	Long:  `Registers the wallet with the blockchain`,
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		if clientWallet == nil {
			fmt.Println("Invalid wallet. Wallet not initialized in sdk")
			return
		}
		wg := &sync.WaitGroup{}
		statusBar := &ZCNStatus{wg: wg}
		wg.Add(1)
		zcncore.RegisterToMiners(clientWallet, statusBar)
		wg.Wait()
		if statusBar.success {
			fmt.Println("Wallet registered")
		} else {
			fmt.Println("Wallet registration failed. " + statusBar.errMsg)
		}
		return
	},
}

func init() {
	rootCmd.AddCommand(registerWalletCmd)
}
