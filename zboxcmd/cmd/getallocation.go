package cmd

import (
	"encoding/json"
	"fmt"

	. "0chain.net/clientsdk/zboxcore/logger"
	"0chain.net/clientsdk/zboxcore/sdk"
	"github.com/spf13/cobra"
)

// getallocationCmd represents the get allocation info command
var getallocationCmd = &cobra.Command{
	Use:   "get",
	Short: "Gets the allocation info",
	Long:  `Gets the allocation info`,
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		fflags := cmd.Flags()                      // fflags is a *flag.FlagSet
		if fflags.Changed("allocation") == false { // check if the flag "path" is set
			fmt.Println("Error: allocation flag is missing") // If not, we'll let the user know
			return                                           // and return
		}
		allocationID := cmd.Flag("allocation").Value.String()
		allocationObj, err := sdk.GetAllocation(allocationID)
		if err != nil {
			Logger.Error("Error fetching the allocation", err)
			fmt.Println("Error fetching/verifying the allocation")
			return
		}
		stats := allocationObj.GetStats()
		statsBytes, _ := json.Marshal(stats)
		fmt.Println(string(statsBytes))
		fmt.Printf("ID : %v\n", allocationObj.ID)
		fmt.Printf("Data Shards  : %v\n", allocationObj.DataShards)
		fmt.Printf("Parity Shards  : %v\n", allocationObj.ParityShards)
		fmt.Printf("Expiration  : %v\n", allocationObj.Expiration)
		fmt.Printf("Blobbers : \n")
		for _, blobber := range allocationObj.Blobbers {
			fmt.Printf("\t%v\n", blobber.Baseurl)
		}

		fmt.Printf("Stats : \n")
		fmt.Printf("\tTotal Size : %v\n", allocationObj.Size)
		fmt.Printf("\tUsed Size : %v\n", allocationObj.Stats.UsedSize)
		fmt.Printf("\tNumber of Writes : %v\n", allocationObj.Stats.NumWrites)
		fmt.Printf("\tTotal Challenges : %v\n", allocationObj.Stats.TotalChallenges)
		fmt.Printf("\tPassed Challenges : %v\n", allocationObj.Stats.SuccessChallenges)
		fmt.Printf("\tFailed Challenges : %v\n", allocationObj.Stats.FailedChallenges)
		fmt.Printf("\tOpen Challenges : %v\n", allocationObj.Stats.OpenChallenges)
		fmt.Printf("\tLast Challenge redeemed : %v\n", allocationObj.Stats.LastestClosedChallengeTxn)
		return
	},
}

func init() {
	rootCmd.AddCommand(getallocationCmd)
	getallocationCmd.PersistentFlags().String("allocation", "", "Allocation ID")
	getallocationCmd.MarkFlagRequired("allocation")
}
