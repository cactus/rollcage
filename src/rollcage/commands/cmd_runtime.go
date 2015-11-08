package commands

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"rollcage/core"

	"github.com/cactus/cobra"
	"github.com/cactus/gologit"
)

func runtimeCmdRun(cmd *cobra.Command, args []string) {
	jailid := core.GetJailUUIDByTagOrUUID(args[0])
	if jailid == "" {
		gologit.Fatalf("No jail found by '%s'\n", args[0])
	}

	out, err := core.Jls("-n", "-j", fmt.Sprintf("ioc-%s", jailid))
	if err != nil {
		if len(out) == 0 || bytes.Contains(out, []byte("not found")) {
			gologit.Fatalf("Jail is not running!\n")
		}
		gologit.Fatalf("Error: %s\n", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), " ")
	for _, line := range lines {
		fmt.Fprintf(os.Stdout, "%s\n", line)
	}
}

func init() {
	cmd := &cobra.Command{
		Use:   "runtime UUID|TAG",
		Short: "show runtime configuration of a jail",
		Long:  "Show runtime configuration of a jail. Useful for debugging.",
		Run:   runtimeCmdRun,
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
