package commands

import (
	"fmt"
	"rollcage/core"
	"strings"
	"time"

	"github.com/cactus/cobra"
	"github.com/cactus/gologit"
)

var recusiveSnapshot bool

func snapshotCmdRun(cmd *cobra.Command, args []string) {
	// requires root
	if !core.IsRoot() {
		gologit.Fatalf("Must be root to snapshot\n")
	}

	jail, err := core.FindJail(args[0])
	if err != nil {
		gologit.Fatalf("No jail found by '%s'\n", args[0])
	}

	var snapname string
	if len(args) > 1 {
		snapname = strings.TrimLeft(args[1], "@")
	} else {
		snapname = fmt.Sprintf(
			"ioc-%s", time.Now().Format("2006-01-02_15:04:05"))
	}

	zfsCmd := []string{"snapshot"}
	if recusiveSnapshot {
		zfsCmd = append(zfsCmd, "-r")
	}
	zfsCmd = append(zfsCmd, fmt.Sprintf("%s/root@%s", jail.Path, snapname))
	core.ZFSMust(fmt.Errorf("Error removing snapshot"), zfsCmd...)
}

func init() {
	cmd := &cobra.Command{
		Use:   "snapshot UUID|TAG snapshotname",
		Short: "Create a zfs snapshot for jail",
		Run:   snapshotCmdRun,
		PreRun: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				gologit.Fatalln("Required UUID|TAG not provided")
			}
		},
	}

	cmd.Flags().BoolVarP(
		&recusiveSnapshot, "recursive", "r", false,
		"do a recursive snapshot of the jail root")

	RootCmd.AddCommand(cmd)
}
