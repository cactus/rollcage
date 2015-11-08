package commands

import (
	"fmt"
	"strings"

	"rollcage/core"

	"github.com/cactus/cobra"
)

func listCmdRun(cmd *cobra.Command, args []string) {
	const propertyList = "org.freebsd.iocage:host_hostuuid," +
		"org.freebsd.iocage:tag,org.freebsd.iocage:boot"
	outputHeaders := []string{"jid", "uuid", "tag", "boot", "state"}

	running := strings.TrimSpace(string(core.JlsMust("jid", "name")))
	jails := make(map[string]string, 0)
	for _, jinfo := range strings.Split(running, "\n") {
		jail := strings.Split(jinfo, " ")
		if len(jail) < 2 {
			continue
		}
		jid := strings.TrimSpace(jail[0])
		jname := strings.TrimSpace(jail[1])
		jails[jname] = jid
	}

	out := core.ZFSMust("list", "-H", "-o", propertyList, "-d", "1", core.GetJailsPath())

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	wf := core.NewOutputWriter(outputHeaders, MachineOutput)
	for _, line := range lines {
		if strings.HasPrefix(line, "-") {
			continue
		}

		iocid := fmt.Sprintf("ioc-%s", strings.SplitN(line, "\t", 2)[0])
		state := "up"
		jid, ok := jails[iocid]
		if !ok {
			jid = "-"
			state = "down"
		}
		fmt.Fprintf(wf, "%s\t%s\t%s\n", jid, line, state)
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
