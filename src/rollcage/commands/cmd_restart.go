package commands

import (
	"fmt"
	"os"
	"os/exec"
	"rollcage/core"
	"time"

	"github.com/cactus/cobra"
	"github.com/cactus/gologit"
)

func restartCmdRun(cmd *cobra.Command, args []string) {
	// requires root
	if !core.IsRoot() {
		gologit.Fatalf("Must be root to stop\n")
	}

	jail, err := core.FindJail(args[0])
	if err != nil {
		gologit.Fatalf("No jail found by '%s'\n", args[0])
	}

	if !jail.IsRunning() {
		gologit.Fatalf("Jail is not running!\n")
	}

	// create file
	f, err := os.OpenFile(jail.GetLogPath(), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		gologit.Fatal(err)
	}
	defer f.Close()

	props := jail.GetProperties()

	jexec := []string{fmt.Sprintf("ioc-%s", jail.HostUUID)}

	jexec_stop := append(jexec, core.SplitFieldsQuoteSafe(props.GetIOC("exec_stop"))...)
	excmd := exec.Command("/usr/sbin/jexec", jexec_stop...)
	excmd.Stdout = f
	excmd.Stderr = f
	err = excmd.Run()
	if err != nil {
		gologit.Printf("%s\n", err)
	}

	jexec_start := append(jexec, core.SplitFieldsQuoteSafe(props.GetIOC("exec_start"))...)
	excmd = exec.Command("/usr/sbin/jexec", jexec_start...)
	excmd.Stdout = f
	excmd.Stderr = f
	err = excmd.Run()
	if err != nil {
		gologit.Printf("%s\n", err)
	}

	// set last_started property
	t := time.Now()
	core.ZFSMust(
		fmt.Errorf("Error setting property"), "set",
		fmt.Sprintf(
			"org.freebsd.iocage:last_started=%s",
			t.Format("2006-01-02_15:04:05")),
		jail.Path)
}

func init() {
	RootCmd.AddCommand(&cobra.Command{
		Use:   "restart UUID|TAG",
		Short: "restart jail",
		Long:  "Restart jail identified by UUID or TAG.",
		Run:   restartCmdRun,
		PreRun: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				gologit.Fatalln("Required UUID|TAG not provided")
			}
		},
	})
}
