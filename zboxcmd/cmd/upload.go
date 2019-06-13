package cmd

import (
	"fmt"
	"sync"

	"github.com/0chain/gosdk/zboxcore/sdk"
	"github.com/spf13/cobra"
)

// uploadCmd represents upload command
var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "upload file to blobbers",
	Long:  `upload file to blobbers`,
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
		if fflags.Changed("localpath") == false {
			fmt.Println("Error: localpath flag is missing")
			return
		}
		allocationID := cmd.Flag("allocation").Value.String()
		allocationObj, err := sdk.GetAllocation(allocationID)
		if err != nil {
			fmt.Println("Error fetching the allocation", err)
			return
		}
		remotepath := cmd.Flag("remotepath").Value.String()
		localpath := cmd.Flag("localpath").Value.String()
		wg := &sync.WaitGroup{}
		statusBar := &StatusBar{wg: wg}
		wg.Add(1)
		allocationObj.UploadFile(localpath, remotepath, statusBar)
		wg.Wait()
		return
	},
}

func init() {
	rootCmd.AddCommand(uploadCmd)
	uploadCmd.PersistentFlags().String("allocation", "", "Allocation ID")
	uploadCmd.PersistentFlags().String("remotepath", "", "Remote path to upload")
	uploadCmd.PersistentFlags().String("localpath", "", "Local path of file to upload")
	uploadCmd.MarkFlagRequired("allocation")
	uploadCmd.MarkFlagRequired("localpath")
	uploadCmd.MarkFlagRequired("remotepath")
}
