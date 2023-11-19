//go:build selfupdate
// +build selfupdate

package main

import (
	"fmt"

	"github.com/mmelnyk/golang-project-layout/internal/selfupdate"
	"github.com/spf13/cobra"
)

var selfupdateCmd = &cobra.Command{
	Use:   "self-update",
	Short: "Self Update operation",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
}

var selfupdateCheckCmd = &cobra.Command{
	Use:          "check",
	Short:        "Check for the latest update",
	Long:         ``,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Current version:   ", buildnumber)
		latest, err := selfupdate.GetLatestVersion(giturl)
		if err == nil {
			fmt.Println("Available version: ", latest)
		}
		return err
	},
}

var selfupdateDownloadCmd = &cobra.Command{
	Use:          "download",
	Short:        "Download the latest update",
	Long:         ``,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := selfupdate.DownloadLatestVersion(giturl, binary, buildnumber)
		return err
	},
}

func init() {
	selfupdateCmd.AddCommand(selfupdateCheckCmd)
	selfupdateCmd.AddCommand(selfupdateDownloadCmd)
	rootCmd.AddCommand(selfupdateCmd)
}
