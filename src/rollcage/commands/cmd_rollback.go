package commands

import (
	"fmt"
	"path"
	"rollcage/core"
	"strings"

	"github.com/cactus/cobra"
	"github.com/cactus/gologit"
)

func rollbackCmdRun(cmd *cobra.Command, args []string) {
	// requires root
	if !core.IsRoot() {
		gologit.Fatalf("Must be root to rollback\n")
	}

	jailpath := core.GetJailByTagOrUUID(args[0])
	if jailpath == "" {
		gologit.Fatalf("No jail found by '%s'\n", args[0])
	}

	snapname := strings.TrimLeft(args[1], "@")

	// get FS's
	lines := core.SplitOutput(
		core.ZFSMust("list", "-Hr", "-o", "name", path.Join(jailpath, "root")))
	if len(lines) < 1 {
		gologit.Fatalf("No datasets at jailpath!\n")
	}

	snapshots := []string{}
	for _, line := range lines {
		out := core.ZFSMust("list", "-Ht", "snapshot", "-o", "name", "-d1",
			fmt.Sprintf("%s@%s", line[0], snapname))
		if len(out) != 0 {
			snapshots = append(snapshots, strings.TrimSpace(string(out)))
		}
	}

	if len(snapshots) == 0 {
		gologit.Fatalln("Snapshot '%s' not found!", snapname)
	}

	for _, snapshot := range snapshots {
		i := strings.LastIndex(snapshot, "@")
		elemName := snapshot[:i]
		j := strings.LastIndex(snapshot, "/")
		elemName = elemName[j:]
		fmt.Printf("* Rolling back jail dataset '%s' to '@%s'\n",
			elemName, snapname)
		core.ZFSMust("rollback", "-r", snapshot)
	}
}

func init() {
	cmd := &cobra.Command{
		Use:   "rollback UUID|TAG snapshotname",
		Short: "Rollback jail to a particular snapshot",
		Run:   rollbackCmdRun,
		PreRun: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				gologit.Fatalln("Required UUID|TAG not provided")
			}
			if len(args) == 1 {
				gologit.Fatalln("Required snapshotname not provided")
			}
		},
	}

	RootCmd.AddCommand(cmd)
}
