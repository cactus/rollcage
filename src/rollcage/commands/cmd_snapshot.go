package commands

import (
	"fmt"
	"rollcage/core"
	"strings"
	"time"

	"github.com/cactus/cobra"
	"github.com/cactus/gologit"
)

func snapshotCmdRun(cmd *cobra.Command, args []string) {
	// requires root
	if !core.IsRoot() {
		gologit.Fatalf("Must be root to snapshot\n")
	}

	jailpath := core.GetJailByTagOrUUID(args[0])
	if jailpath == "" {
		gologit.Fatalf("No jail found by '%s'\n", args[0])
	}

	var snapname string
	if len(args) > 1 {
		snapname = strings.TrimLeft(args[1], "@")
	} else {
		snapname = fmt.Sprintf(
			"ioc-%s", time.Now().Format("2006-01-02_15:04:05"))
	}

	core.ZFSMust("snapshot", fmt.Sprintf("%s/root@%s", jailpath, snapname))
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

	RootCmd.AddCommand(cmd)
}
