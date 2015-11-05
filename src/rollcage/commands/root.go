package commands

import (
	"log"
	"rollcage/core"

	"github.com/cactus/cobra"
	"github.com/cactus/gologit"
)

var Verbose, MachineOutput, ParsableValues bool
var ConfigPath string
var RootCmd = &cobra.Command{
	Use:   "rollcage [sub]",
	Short: "rollcage jail manager",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// set logger debug level and start toggle on signal handler
		logger := gologit.Logger
		logger.Set(Verbose)
		// required to update after setting verbose or gologit overwrites
		if Verbose {
			logger.SetFlags(log.Lshortfile)
		} else {
			logger.SetFlags(0)
		}
		logger.Debugln("Debug logging enabled")
		if cmd.Name() != "version" {
			// load config
			core.LoadConfig(ConfigPath)
		}
	},
}

func init() {
	RootCmd.PersistentFlags().StringVarP(
		&ConfigPath, "config", "c", "/usr/local/etc/rollcage.conf",
		"config path")
	RootCmd.PersistentFlags().BoolVarP(
		&Verbose, "verbose", "v", false,
		"turn on verbose logging")
	RootCmd.PersistentFlags().BoolVarP(
		&MachineOutput, "no-headers", "H", false,
		"No headers and tab delimited fields (for scripting).")
}
