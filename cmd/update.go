package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Check for available updates for the ao client, and downloads the update if available.",
	Long:  `Available updates are searched for using a service in the OpenShift cluster.`,
	RunE:  Update,
}

func init() {
	if runtime.GOOS != "windows" {
		RootCmd.AddCommand(updateCmd)
	}
}

func Update(cmd *cobra.Command, args []string) error {
	err := AO.Update(true)
	if err != nil {
		return err
	} else {
		fmt.Println("AO has been updated")
	}
	return nil
}
