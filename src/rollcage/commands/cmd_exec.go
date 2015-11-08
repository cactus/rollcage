package commands

import (
	"fmt"
	"os"
	"syscall"

	"rollcage/core"

	"github.com/cactus/cobra"
	"github.com/cactus/gologit"
)

func execCmdRun(cmd *cobra.Command, args []string) {
	// requires root
	if !core.IsRoot() {
		gologit.Fatalf("Must be root to use exec\n")
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
	execFib := lines[0][1]

	jexec := []string{}
	if execFib != "0" {
		jexec = append(jexec, "/usr/sbin/setfib", execFib)
	}
	jexec = append(jexec, "/usr/sbin/jexec")
	if hostUser != "" {
		jexec = append(jexec, "-u", hostUser)
	}
	if jailUser != "" {
		jexec = append(jexec, "-U", jailUser)
	}
	jexec = append(jexec, fmt.Sprintf("ioc-%s", jail.HostUUID))
	jexec = append(jexec, args[1:]...)

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

var hostUser, jailUser string

func init() {
	cmd := &cobra.Command{
		Use:   "exec [-u username] [-U username] UUID|TAG COMMAND",
		Short: "Execute login to have a shell inside the jail.",
		Run:   execCmdRun,
		PreRun: func(cmd *cobra.Command, args []string) {
			if hostUser != "" && jailUser != "" {
				gologit.Fatalln("Cannot supply both -u and -U")
			}
			arglen := len(args)
			if arglen < 1 {
				gologit.Fatalln("Required UUID|TAG not provided")
			}
			if arglen < 2 {
				gologit.Fatalln("Required command not provided")
			}
		},
	}

	cmd.Flags().StringVarP(
		&hostUser, "host-user", "u", "",
		"user name from host environment as whom the command should run")
	cmd.Flags().StringVarP(
		&jailUser, "jail-user", "U", "",
		"user name from jailed environment as whom the command should run")

	RootCmd.AddCommand(cmd)
}
