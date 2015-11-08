package commands

import (
	"fmt"
	"regexp"
	"rollcage/core"
	"strings"

	"github.com/cactus/cobra"
	"github.com/cactus/gologit"
)

var snapremoveRegex bool

func snapremoveCmdRun(cmd *cobra.Command, args []string) {
	// requires root
	if !core.IsRoot() {
		gologit.Fatalf("Must be root to snapremove\n")
	}

	jail, err := core.FindJail(args[0])
	if err != nil {
		gologit.Fatalf("No jail found by '%s'\n", args[0])
	}

	matchers := args[1:]
	gologit.Debugf("matchers: %#v\n", matchers)

	zfsArgs := []string{"list", "-Hrt", "snapshot",
		"-o", "name", "-d2", jail.Path}
	lines := core.SplitOutput(core.ZFSMust(zfsArgs...))

	rmlist := []string{}
	for _, line := range lines {
		if len(line) == 0 || len(line[0]) == 0 {
			continue
		}
		snapname := strings.SplitN(line[0], "@", 2)[1]
		gologit.Debugf("source snapname: %#v\n", snapname)
		for _, m := range matchers {
			if snapremoveRegex {
				matched, err := regexp.MatchString(m, snapname)
				if err != nil {
					gologit.Fatalf("Regex error: %s", err)
				}
				if matched {
					rmlist = append(rmlist, line[0])
					continue
				}
			} else {
				if m == snapname {
					rmlist = append(rmlist, line[0])
					continue
				}
			}
		}
	}
	gologit.Debugf("match list: %#v\n", rmlist)

	for _, snap := range rmlist {
		fmt.Printf("Removing snapshot: %s\n", strings.SplitN(snap, "@", 2)[1])
		core.ZFSMust("destroy", "-r", snap)
	}
}

func init() {
	cmd := &cobra.Command{
		Use:   "snapremove UUID|TAG snapshotname [snapshotname ...]",
		Short: "Remove snapshots belonging to jail",
		Run:   snapremoveCmdRun,
		PreRun: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				gologit.Fatalln("Required UUID|TAG not provided")
			} else if len(args) == 1 {
				gologit.Fatalln("Required snapshotname not provided")
			}
		},
	}

	cmd.Flags().BoolVarP(
		&snapremoveRegex, "regex", "x", false,
		"snapshotname becomes a match regex")

	RootCmd.AddCommand(cmd)
}
