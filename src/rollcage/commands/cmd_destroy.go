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

var force bool = false

func destroyCmdRun(cmd *cobra.Command, args []string) {
	// requires root
	if !core.IsRoot() {
		gologit.Fatalf("Must be root to destroy\n")
	}

	var jailpath, jailUUID string
	for _, line := range core.SplitOutput(core.ZFSMust(
		"list", "-H", "-d", "1",
		"-o", "name,org.freebsd.iocage:host_hostuuid,org.freebsd.iocage:tag",
		core.GetJailsPath())) {
		if line[2] == args[0] || strings.HasPrefix(line[1], args[0]) {
			jailpath = line[0]
			jailUUID = line[1]
			break
		}
	}
	if jailpath == "" {
		gologit.Fatalf("Jail '%s' not found!\n", args[0])
	}

	out, err := core.Jls("-j", fmt.Sprintf("ioc-%s", jailUUID), "jid")
	if err == nil && !bytes.Contains(out, []byte("not found")) {
		gologit.Fatalf("Jail is running. Shutdown first.\n")
	}

	propertyList := []string{
		"mountpoint",
		"org.freebsd.iocage:type",
		"org.freebsd.iocage:tag",
	}

	lines := core.SplitOutput(core.ZFSMust("list", "-H", "-o", strings.Join(propertyList, ","), jailpath))
	if len(lines) < 1 {
		gologit.Fatalf("No output from property fetch\n")
	}

	prop_mountpoint := removeDash(lines[0][0])
	prop_type := removeDash(lines[0][1])
	prop_tag := removeDash(lines[0][2])

	if prop_type != "thickjail" {
		gologit.Fatalf("Type is not thickjail.\nI don't know how to handle this yet.\nGiving up!")
	}

	fmt.Print("Are you sure [yN]? :")
	var response string
	_, err = fmt.Scanln(&response)
	if err != nil {
		gologit.Fatalf("%s", err)
	}

	if !force {
		response = strings.ToLower(strings.TrimSpace(response))
		if len(response) != 1 || response[0] != 'y' {
			return
		}
	}

	fmt.Printf("Destroying: %s (%s)\n", jailUUID, prop_tag)
	core.ZFSMust("destroy", "-fr", jailpath)
	os.RemoveAll(prop_mountpoint)
}

func init() {
	cmd := &cobra.Command{
		Use:   "destroy UUID|TAG",
		Short: "destroy jail",
		Long:  "destroy jail identified by UUID or TAG.",
		Run:   destroyCmdRun,
		PreRun: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				gologit.Fatalln("Required UUID|TAG not provided")
			}
		},
	}
	cmd.Flags().BoolVarP(
		&force, "force", "f",
		false, "attempt to remove jail without prompting for confirmation")
	RootCmd.AddCommand(cmd)
}
