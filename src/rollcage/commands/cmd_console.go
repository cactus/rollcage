package commands

import (
	"fmt"
	"os"
	"os/exec"

	"rollcage/core"

	"github.com/cactus/cobra"
	"github.com/cactus/gologit"
)

func consoleCmdRun(cmd *cobra.Command, args []string) {
	// requires root
	if !core.IsRoot() {
		gologit.Fatalf("Must be root to use console\n")
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
	loginFlags := lines[0][0]
	execFib := lines[0][1]

	jexec := []string{}
	if execFib != "0" {
		jexec = append(jexec, "/usr/sbin/setfib", execFib)
	}
	jexec = append(jexec, "/usr/sbin/jexec", fmt.Sprintf("ioc-%s", jailUUID), "login")
	jexec = append(jexec, core.SplitFieldsQuoteSafe(loginFlags)...)

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
