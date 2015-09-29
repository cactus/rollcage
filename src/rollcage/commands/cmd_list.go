package commands

import (
	"bytes"
	"fmt"
	"strings"

	"rollcage/core"

	"github.com/cactus/cobra"
)

func listCmdRun(cmd *cobra.Command, args []string) {
	const propertyList = "org.freebsd.iocage:host_hostuuid," +
		"org.freebsd.iocage:tag,org.freebsd.iocage:boot"
	outputHeaders := []string{"jid", "uuid", "tag", "boot", "state"}

	out := core.ZFSMust("list", "-H", "-o", propertyList, "-d", "1", core.GetJailsPath())

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	wf := core.NewOutputWriter(outputHeaders, MachineOutput)
	for _, line := range lines {
		if strings.HasPrefix(line, "-") {
			continue
		}
		hostUUID := strings.SplitN(line, "\t", 2)[0]
		jid := core.JlsMust("-j", fmt.Sprintf("ioc-%s", hostUUID), "jid")
		state := "up"
		if string(jid) == "" {
			jid = []byte("-")
			state = "down"
		}
		fmt.Fprintf(wf, "%s\t%s\t%s\n", bytes.TrimSpace(jid), line, state)
	}
	wf.Flush()
}

func init() {
	RootCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all jails",
		Run:   listCmdRun,
	})
}
