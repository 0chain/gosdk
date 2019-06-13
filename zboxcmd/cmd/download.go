package cmd

import (
	"fmt"
	"sync"

	"github.com/0chain/gosdk/zboxcore/sdk"
	"github.com/spf13/cobra"
)

// downloadCmd represents download command
var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "download file from blobbers",
	Long:  `download file from blobbers`,
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		fflags := cmd.Flags() // fflags is a *flag.FlagSet
		if fflags.Changed("remotepath") == false && fflags.Changed("authticket") == false {
			fmt.Println("Error: remotepath / authticket flag is missing")
			return
		}

		remotepath := cmd.Flag("remotepath").Value.String()
		authticket := cmd.Flag("authticket").Value.String()
		if len(remotepath) == 0 && len(authticket) == 0 {
			fmt.Println("Error: remotepath / authticket flag is missing")
			return
		}

		localpath := cmd.Flag("localpath").Value.String()
		wg := &sync.WaitGroup{}
		statusBar := &StatusBar{wg: wg}
		wg.Add(1)
		var err error
		if len(remotepath) > 0 {
			if fflags.Changed("allocation") == false { // check if the flag "path" is set
				fmt.Println("Error: allocation flag is missing") // If not, we'll let the user know
				return                                           // and return
			}
			allocationID := cmd.Flag("allocation").Value.String()
			allocationObj, err := sdk.GetAllocation(allocationID)
			if err != nil {
				fmt.Println("Error fetching the allocation", err)
				return
			}
			err = allocationObj.DownloadFile(localpath, remotepath, statusBar)
		} else if len(authticket) > 0 {
			allocationObj, err := sdk.GetAllocationFromAuthTicket(authticket)
			if err != nil {
				fmt.Println("Error fetching the allocation", err)
				return
			}
			err = allocationObj.DownloadFromAuthTicket(localpath, authticket, statusBar)
		}

		if err == nil {
			wg.Wait()
		} else {
			fmt.Println(err.Error())
		}
		return
	},
}

func init() {
	rootCmd.AddCommand(downloadCmd)
	downloadCmd.PersistentFlags().String("allocation", "", "Allocation ID")
	downloadCmd.PersistentFlags().String("remotepath", "", "Remote path to download")
	downloadCmd.PersistentFlags().String("localpath", "", "Local path of file to download")
	downloadCmd.PersistentFlags().String("authticket", "", "Auth ticket fot the file to download if you dont own it")
	downloadCmd.MarkFlagRequired("allocation")
	downloadCmd.MarkFlagRequired("localpath")
}
