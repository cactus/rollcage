package commands

import (
	"fmt"
	"strings"

	"rollcage/core"

	"github.com/cactus/cobra"
	"github.com/cactus/gologit"
)

func snaplistCmdRun(cmd *cobra.Command, args []string) {
	jailpath := core.GetJailByTagOrUUID(args[0])
	if jailpath == "" {
		gologit.Fatalf("No jail found by '%s'\n", args[0])
	}

	zfsArgs := []string{"list", "-Hrt", "snapshot",
		"-o", "name,creation,used,referenced", "-d2"}
	if ParsableValues {
		zfsArgs = append(zfsArgs, "-p")
	}
	zfsArgs = append(zfsArgs, jailpath)

	lines := core.SplitOutput(core.ZFSMust(zfsArgs...))
	gologit.Debugf("%#v", lines)
	if len(lines) == 0 || len(lines[0]) == 0 || len(lines[0][0]) == 0 {
		return
	}

	outputHeaders := []string{"name", "created", "rsize", "used"}
	wf := core.NewOutputWriter(outputHeaders, MachineOutput)
	for _, line := range lines {
		if len(line) < 4 {
			continue
		}

		snapname := strings.SplitN(line[0], "@", 2)[1]
		fmt.Fprintf(wf, "%s\t%s\t%s\t%s\n", snapname, line[1], line[2], line[3])
	}
	wf.Flush()
}

func init() {
	cmd := &cobra.Command{
		Use:   "snaplist UUID|TAG [command]",
		Short: "List all snapshots belonging to jail",
		Run:   snaplistCmdRun,
		PreRun: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				gologit.Fatalln("Required UUID|TAG not provided")
			}
		},
	}

	cmd.Flags().BoolVarP(
		&ParsableValues, "parsable-values", "p", false,
		"output parsable (exact) values")

	RootCmd.AddCommand(cmd)
}
