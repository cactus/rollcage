package commands

import (
	"fmt"
	"strings"

	"rollcage/core"

	"github.com/cactus/cobra"
	"github.com/cactus/gologit"
)

func dfCmdRun(cmd *cobra.Command, args []string) {
	propertyList := "org.freebsd.iocage:host_hostuuid," +
		"org.freebsd.iocage:tag,compressratio,reservation," +
		"quota,used,available"
	outputHeaders := []string{"uuid", "tag", "crt", "res", "qta", "use", "ava"}

	zfsArgs := []string{"list", "-H", "-o", propertyList}
	if ParsableValues {
		zfsArgs = append(zfsArgs, "-p")
	}
	if len(args) == 0 {
		zfsArgs = append(zfsArgs, "-d", "1", core.GetJailsPath())
	} else {
		jail, err := core.FindJail(args[0])
		if err != nil {
			gologit.Fatalf("No jail found by '%s'\n", args[0])
		}
		zfsArgs = append(zfsArgs, jail.Path)
	}
	out := core.ZFSMust(zfsArgs...)
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	wf := core.NewOutputWriter(outputHeaders, MachineOutput)
	for _, line := range lines {
		if strings.HasPrefix(line, "-") {
			continue
		}
		fmt.Fprintf(wf, "%s\n", line)
	}
	wf.Flush()
}

const dfLongHelp = `
List disk space related information.

Available fields:
  UUID - jail UUID
  TAG  - jail name
  CRT  - compression ratio
  RES  - reserved space
  QTA  - disk quota
  USE  - used space
  AVA  - available space`

func init() {
	cmd := &cobra.Command{
		Use:   "df [UUID|TAG]",
		Short: "List disk space related information",
		Long:  dfLongHelp,
		Run:   dfCmdRun,
	}

	cmd.Flags().BoolVarP(
		&ParsableValues, "parsable-values", "p", false,
		"output parsable (exact) values")

	RootCmd.AddCommand(cmd)
}
