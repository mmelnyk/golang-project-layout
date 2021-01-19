package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mmelnyk/golang-project-layout/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Config information",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
}

var configDefaultCmd = &cobra.Command{
	Use:   "default",
	Short: "Dump default config structure",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := &appconfig{}
		cfg.Log.Reset()
		cfg.Lib.Reset()
		if dump, err := config.Marshal(cfg, appshortname, currentConfigVersion); err == nil {
			fmt.Println("# Default configuration for", filepath.Base(os.Args[0]))
			fmt.Println(string(dump))
		}
	},
}

var configDumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Dump config structure",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := getConfig()

		// Cleanup
		if err == nil {
			defer cfg.cleanup()
		}

		if err == nil {
			if dump, err := config.Marshal(cfg, appshortname, currentConfigVersion); err == nil {
				fmt.Println("# Used configuration for", filepath.Base(os.Args[0]))
				fmt.Println(string(dump))
			}
		}

		// Suppress usage message (we showed all required messages,
		// so just allow to return status code to OS)
		cmd.SilenceUsage = true

		return err
	},
}

var configCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Do validation check for the config file",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := getConfig()

		// Cleanup
		if err == nil {
			defer cfg.cleanup()
		}

		// Add extra validation if needed
		err = cfg.check()

		message := "Config file validation check passed"
		if err != nil {
			message = "Config file validation check failed"
		}
		fmt.Println(message)

		// Suppress usage message (we showed all required messages,
		// so just allow to return status code to OS)
		cmd.SilenceUsage = true

		return err
	},
}

func init() {
	configCmd.AddCommand(configDefaultCmd)
	configCmd.AddCommand(configCheckCmd)
	configCmd.AddCommand(configDumpCmd)
	rootCmd.AddCommand(configCmd)
}
