package cmd

import (
	"fmt"
	"os"

	"0chain.net/clientsdk/core/common"
	"0chain.net/clientsdk/zboxcore/sdk"
	"github.com/spf13/cobra"
)

var datashards, parityshards *int
var size *int64

// newallocationCmd represents the new allocation command
var newallocationCmd = &cobra.Command{
	Use:   "newallocation",
	Short: "Creates a new allocation",
	Long:  `Creates a new allocation`,
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		if datashards == nil || parityshards == nil || size == nil {
			fmt.Println("Invalid allocation parameters.")
			os.Exit(1)
		}
		allocationID, err := sdk.CreateAllocation(*datashards, *parityshards, *size, common.Now()+7776000)
		if err != nil {
			fmt.Println("Error creating allocation." + err.Error())
			os.Exit(1)
		}
		fmt.Println("Allocation created : " + allocationID)
		return
	},
}

func init() {
	rootCmd.AddCommand(newallocationCmd)
	datashards = newallocationCmd.PersistentFlags().Int("data", 2, "--data 2")
	parityshards = newallocationCmd.PersistentFlags().Int("parity", 2, "--parity 2")
	size = newallocationCmd.PersistentFlags().Int64("size", 2147483648, "--size 10000")
	newallocationCmd.MarkFlagRequired("data")
	newallocationCmd.MarkFlagRequired("parity")
	newallocationCmd.MarkFlagRequired("size")
}
