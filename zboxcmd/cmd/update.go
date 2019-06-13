package cmd

import (
	"fmt"
	"sync"

	"github.com/0chain/gosdk/zboxcore/sdk"
	"github.com/spf13/cobra"
)

// updateCmd represents update file command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update file to blobbers",
	Long:  `update file to blobbers`,
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		fflags := cmd.Flags()
		if fflags.Changed("allocation") == false {
			fmt.Println("Error: allocation flag is missing")
			return
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
		allocationObj.UpdateFile(localpath, remotepath, statusBar)
		wg.Wait()
		return
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.PersistentFlags().String("allocation", "", "Allocation ID")
	updateCmd.PersistentFlags().String("remotepath", "", "Remote path to upload")
	updateCmd.PersistentFlags().String("localpath", "", "Local path of file to upload")
	updateCmd.MarkFlagRequired("allocation")
	updateCmd.MarkFlagRequired("localpath")
	updateCmd.MarkFlagRequired("remotepath")
}
