package commands

import (
	"fmt"
	"os"
	"os/exec"

	"rollcage/core"

	"github.com/cactus/cobra"
	"github.com/cactus/gologit"
)

func execCmdRun(cmd *cobra.Command, args []string) {
	// requires root
	if !core.IsRoot() {
		gologit.Fatalf("Must be root to use exec\n")
	}

	jailUUID := core.GetJailUUIDByTagOrUUID(args[0])
	if jailUUID == "" {
		gologit.Fatalf("No jail found by '%s'\n", args[0])
	}

	jid := string(core.JlsMust("-j", fmt.Sprintf("ioc-%s", jailUUID), "jid"))
	if jid == "" {
		gologit.Fatalf("Jail is not running!\n")
	}

	jailpath := core.GetJailByTagOrUUID(args[0])
	if jailpath == "" {
		gologit.Fatalf("No jail found by '%s'\n", args[0])
	}

	// get exec fib property
	lines := core.SplitOutput(core.ZFSMust("list", "-H", "-o", "org.freebsd.iocage:login_flags,org.freebsd.iocage:exec_fib", jailpath))
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
	jexec = append(jexec, fmt.Sprintf("ioc-%s", jailUUID))
	jexec = append(jexec, args[1:]...)

	excmd := exec.Command(jexec[0], jexec[1:]...)
	excmd.Env = []string{
		"PATH=/sbin:/bin:/usr/sbin:/usr/bin:/usr/local/sbin:/usr/local/bin",
		fmt.Sprintf("TERM=%s", os.Getenv("TERM")),
	}
	gologit.Debugf("%s %#s\n", excmd.Path, excmd.Args)
	err := excmd.Run()
	if err != nil {
		gologit.Fatal(err)
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
