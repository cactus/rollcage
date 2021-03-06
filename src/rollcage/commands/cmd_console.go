package commands

import (
	"fmt"
	"os"
	"syscall"

	"rollcage/core"

	"github.com/cactus/cobra"
	"github.com/cactus/gologit"
)

func consoleCmdRun(cmd *cobra.Command, args []string) {
	// requires root
	if !core.IsRoot() {
		gologit.Fatalf("Must be root to use console\n")
	}

	jail, err := core.FindJail(args[0])
	if err != nil {
		gologit.Fatalf("No jail found by '%s'\n", args[0])
	}

	if !jail.IsRunning() {
		gologit.Fatalf("Jail is not running!\n")
	}

	// get exec fib property
	lines := core.SplitOutput(core.ZFSMust(
		fmt.Errorf("Error listing jails"),
		"list", "-H",
		"-o", "org.freebsd.iocage:login_flags,org.freebsd.iocage:exec_fib",
		jail.Path))
	loginFlags := lines[0][0]
	execFib := lines[0][1]

	jexec := []string{}
	if execFib != "0" {
		jexec = append(jexec, "/usr/sbin/setfib", execFib)
	}
	jexec = append(jexec, "/usr/sbin/jexec", fmt.Sprintf("ioc-%s", jail.HostUUID), "login")
	jexec = append(jexec, core.SplitFieldsQuoteSafe(loginFlags)...)

	// set a default path
	environ := []string{
		"PATH=/sbin:/bin:/usr/sbin:/usr/bin:/usr/local/sbin:/usr/local/bin",
	}
	// set a term from caller
	environ = append(environ, fmt.Sprintf("TERM=%s", os.Getenv("TERM")))

	gologit.Debugf("%#s\n", jexec)
	execErr := syscall.Exec(jexec[0], jexec, environ)
	if execErr != nil {
		gologit.Fatal(execErr)
	}
}

func init() {
	RootCmd.AddCommand(&cobra.Command{
		Use:   "console UUID|TAG",
		Short: "Execute login to have a shell inside the jail.",
		Run:   consoleCmdRun,
		PreRun: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				gologit.Fatalln("Required UUID|TAG not provided")
			}
		},
	})
}
