package commands

import (
	"fmt"
	"os"
	"path"
	"syscall"

	"rollcage/core"

	"github.com/cactus/cobra"
	"github.com/cactus/gologit"
)

func chrootCmdRun(cmd *cobra.Command, args []string) {
	// requires root
	if !core.IsRoot() {
		gologit.Fatalf("Must be root to chroot\n")
	}

	jail, err := core.FindJail(args[0])
	if err != nil {
		gologit.Fatalf("No jail found by '%s'\n", args[0])
	}
	propertyOut := core.ZFSMust(
		fmt.Errorf("Error getting properties"),
		"get", "-H", "-o", "value", "mountpoint", jail.Path)

	chrootArgs := []string{
		"/usr/sbin/chroot",
		path.Join(propertyOut, "root"),
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

	// set a default path
	environ := []string{
		"PATH=/sbin:/bin:/usr/sbin:/usr/bin:/usr/local/sbin:/usr/local/bin",
	}
	// set a term from caller
	environ = append(environ, fmt.Sprintf("TERM=%s", os.Getenv("TERM")))

	execErr := syscall.Exec(chrootArgs[0], chrootArgs, environ)
	if execErr != nil {
		gologit.Fatal(execErr)
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
