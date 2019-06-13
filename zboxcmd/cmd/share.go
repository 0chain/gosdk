package cmd

import (
	"fmt"

	"0chain.net/clientsdk/zboxcore/sdk"
	"github.com/spf13/cobra"
)

// shareCmd represents share command
var shareCmd = &cobra.Command{
	Use:   "share",
	Short: "share files from blobbers",
	Long:  `share files from blobbers`,
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
		ref, err := allocationObj.GetAuthTicketForShare(remotepath, "")
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println("Auth token :" + ref)

		return
	},
}

func init() {
	rootCmd.AddCommand(shareCmd)
	shareCmd.PersistentFlags().String("allocation", "", "Allocation ID")
	shareCmd.PersistentFlags().String("remotepath", "", "Remote path to share")
	shareCmd.MarkFlagRequired("allocation")
	shareCmd.MarkFlagRequired("remotepath")
}
