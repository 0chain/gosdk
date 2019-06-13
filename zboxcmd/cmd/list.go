package cmd

import (
	"fmt"
	"os"
	"strconv"

	"0chain.net/clientsdk/zboxcmd/util"
	"0chain.net/clientsdk/zboxcore/fileref"
	"0chain.net/clientsdk/zboxcore/sdk"
	"github.com/spf13/cobra"
)

// listCmd represents list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list files from blobbers",
	Long:  `list files from blobbers`,
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
		ref, err := allocationObj.ListDir(remotepath)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		header := []string{"Type", "Name", "Path", "Size", "Num Blocks"}
		data := make([][]string, len(ref.Children))
		for idx, child := range ref.Children {
			size := strconv.FormatInt(child.Size, 10)
			if child.Type == fileref.DIRECTORY {
				size = ""
			}
			data[idx] = []string{child.Type, child.Name, child.Path, size, strconv.FormatInt(child.NumBlocks, 10)}
		}
		util.WriteTable(os.Stdout, header, []string{}, data)
		return
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.PersistentFlags().String("allocation", "", "Allocation ID")
	listCmd.PersistentFlags().String("remotepath", "", "Remote path to list from")
	listCmd.MarkFlagRequired("allocation")
	listCmd.MarkFlagRequired("remotepath")
}
