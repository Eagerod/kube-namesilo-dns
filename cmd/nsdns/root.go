package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "nsdns",
	Short: "pull ingress resources, and set DNS according to the host's external DNS",
}

func Execute() {
	rootCmd.AddCommand(updateCommand())
	rootCmd.AddCommand(versionCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
