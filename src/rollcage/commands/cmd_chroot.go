package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"rollcage/core"

	"github.com/cactus/cobra"
	"github.com/cactus/gologit"
)

func chrootCmdRun(cmd *cobra.Command, args []string) {
	// requires root
	if !core.IsRoot() {
		gologit.Fatalf("Must be root to chroot\n")
	}

	jailpath := core.GetJailByTagOrUUID(args[0])
	if jailpath == "" {
		gologit.Fatalf("No jail found by '%s'\n", args[0])
	}
	propertyOut := core.ZFSMust("get", "-H", "-o", "value", "mountpoint", jailpath)

	chrootArgs := []string{
		"/usr/sbin/chroot",
		path.Join(strings.TrimSpace(string(propertyOut)), "root"),
	}

	if len(args) > 1 {
		chrootArgs = append(chrootArgs, args[1:]...)
	} else {
		shell := os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/sh"
		}
		chrootArgs = append(chrootArgs, shell)
	}

	excmd := exec.Command(chrootArgs[0], chrootArgs[1:]...)
	excmd.Env = []string{
		"PATH=/sbin:/bin:/usr/sbin:/usr/bin:/usr/local/sbin:/usr/local/bin",
		fmt.Sprintf("TERM=%s", os.Getenv("TERM")),
	}
	gologit.Debugln("%#s\n", excmd.Args)
	err := excmd.Run()
	if err != nil {
		gologit.Fatal(err)
	}
}

func init() {
	RootCmd.AddCommand(&cobra.Command{
		Use:   "chroot UUID|TAG [command]",
		Short: "Chroot into jail, without actually starting the jail itself",
		Long: `
Chroot into jail, without actually starting the jail itself.

Useful for initial setup (set root password, configure networking).
You can specify a command just like with the normal system chroot tool.`,
		Run: chrootCmdRun,
		PreRun: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				gologit.Fatalln("Required UUID|TAG not provided")
			}
		},
	})
}
