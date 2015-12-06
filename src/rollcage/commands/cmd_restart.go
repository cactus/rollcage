package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path"
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

	// get log dir
	logdir := core.ZFSMust(
		fmt.Errorf("Error setting property"),
		"get", "-H", "-o", "value", "mountpoint",
		path.Join(core.GetZFSRootPath(), "log"))
	logpath := path.Join(logdir, fmt.Sprintf("%s-console.log", jail.HostUUID))

	// create file
	f, err := os.OpenFile(logpath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		gologit.Fatal(err)
	}
	f.WriteString("Starting...\n")
	f.Close()

	props := jail.GetProperties()

	jexec := []string{fmt.Sprintf("ioc-%s", jail.HostUUID)}

	jexec_stop := append(jexec, core.SplitFieldsQuoteSafe(props.GetIOC("exec_stop"))...)
	out, err := exec.Command("/usr/sbin/jexec", jexec_stop...).CombinedOutput()
	gologit.Debugln(string(out))
	f.Write(out)
	if err != nil {
		gologit.Printf("%s\n", err)
	}

	jexec_start := append(jexec, core.SplitFieldsQuoteSafe(props.GetIOC("exec_start"))...)
	out, err = exec.Command("/usr/sbin/jexec", jexec_start...).CombinedOutput()
	gologit.Debugln(string(out))
	f.Write(out)
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
