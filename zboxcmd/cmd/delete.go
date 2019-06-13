package cmd

import (
	"fmt"

	"0chain.net/clientsdk/zboxcore/sdk"
	"github.com/spf13/cobra"
)

// deleteCmd represents delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete file from blobbers",
	Long:  `delete file from blobbers`,
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
		err = allocationObj.DeleteFile(remotepath)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println(remotepath + " deleted")
		return
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.PersistentFlags().String("allocation", "", "Allocation ID")
	deleteCmd.PersistentFlags().String("remotepath", "", "Remote path of file to delete")
	deleteCmd.MarkFlagRequired("allocation")
	deleteCmd.MarkFlagRequired("remotepath")
}
