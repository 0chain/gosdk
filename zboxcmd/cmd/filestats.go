package cmd

import (
	"fmt"
	"os"
	"strconv"

	"0chain.net/clientsdk/zboxcmd/util"
	"0chain.net/clientsdk/zboxcore/sdk"
	"github.com/spf13/cobra"
)

// statsCmd represents list command
var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "stats for file from blobbers",
	Long:  `stats for file from blobbers`,
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		fflags := cmd.Flags()                      // fflags is a *flag.FlagSet
		if fflags.Changed("allocation") == false { // check if the flag "path" is set
			fmt.Println("Error: allocation flag is missing") // If not, we'll let the user know
			return                                           // and return
		}
		if fflags.Changed("remotepath") == false {
			fmt.Println("Error: remotepath flag is missing")
			return
		}
		allocationID := cmd.Flag("allocation").Value.String()
		allocationObj, err := sdk.GetAllocation(allocationID)
		if err != nil {
			fmt.Println("Error fetching the allocation", err)
			return
		}
		remotepath := cmd.Flag("remotepath").Value.String()
		ref, err := allocationObj.GetFileStats(remotepath)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		header := []string{"Blobber", "Name", "Path", "Size", "Uploads", "Block Downloads", "Challenges", "Blockchain Aware"}
		data := make([][]string, 0)
		idx := 0
		for k, v := range ref {
			size := strconv.FormatInt(v.Size, 10)
			rowData := []string{k, v.Name, v.Path, size, strconv.FormatInt(v.NumUpdates, 10), strconv.FormatInt(v.NumBlockDownloads, 10), strconv.FormatInt(v.SuccessChallenges, 10), strconv.FormatBool(len(v.WriteMarkerRedeemTxn) > 0)}
			data = append(data, rowData)
			idx++
		}

		util.WriteTable(os.Stdout, header, []string{}, data)
		return
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)
	statsCmd.PersistentFlags().String("allocation", "", "Allocation ID")
	statsCmd.PersistentFlags().String("remotepath", "", "Remote path to list from")
	statsCmd.MarkFlagRequired("allocation")
	statsCmd.MarkFlagRequired("remotepath")
}
