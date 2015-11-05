package commands

import (
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
		gologit.Fatalf("Must be root to destroy\n")
	}

	jailpath := core.GetJailByTagOrUUID(args[0])
	if jailpath == "" {
		gologit.Fatalf("No jail found by '%s'\n", args[0])
	}

	matchers := args[1:]

	zfsArgs := []string{"list", "-Hrt", "snapshot",
		"-o", "name", "-d2", jailpath}
	lines := core.SplitOutput(core.ZFSMust(zfsArgs...))

	rmlist := []string{}
	for _, l := range lines {
		line := l[0]
		for _, m := range matchers {
			if snapremoveRegex {
				matched, err := regexp.MatchString(m, line)
				if err != nil {
					gologit.Fatalf("Regex error: %s", err)
				}
				if matched {
					rmlist = append(rmlist, line)
					continue
				}
			} else {
				if m == line {
					rmlist = append(rmlist, line)
					continue
				}
			}
		}
	}
	gologit.Debugf("match list: %#v\n", rmlist)

	for _, snap := range rmlist {
		gologit.Printf("Removing snapshot: %s", strings.Split(snap, "@")[1])
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
		&snapremoveRegex, "regex", "r", false,
		"snapshotname becomes a match regex")

	RootCmd.AddCommand(cmd)
}
