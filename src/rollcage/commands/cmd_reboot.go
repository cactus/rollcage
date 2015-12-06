package commands

import (
	"rollcage/core"

	"github.com/cactus/cobra"
	"github.com/cactus/gologit"
)

func rebootCmdRun(cmd *cobra.Command, args []string) {
	// requires root
	if !core.IsRoot() {
		gologit.Fatalf("Must be root to stop\n")
	}

	jail, err := core.FindJail(args[0])
	if err != nil {
		gologit.Fatalf("No jail found by '%s'\n", args[0])
	}

	if !jail.IsRunning() {
		gologit.Fatalf("Jail is not running!\n")
	}

	stopCmdRun(cmd, args)
	startCmdRun(cmd, args)
}

func init() {
	RootCmd.AddCommand(&cobra.Command{
		Use:   "reboot UUID|TAG",
		Short: "reboot jail",
		Long:  "Reboot jail identified by UUID or TAG.",
		Run:   rebootCmdRun,
		PreRun: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				gologit.Fatalln("Required UUID|TAG not provided")
			}
		},
	})
}
