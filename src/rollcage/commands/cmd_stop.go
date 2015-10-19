package commands

import (
	"fmt"
	"os/exec"
	"path"
	"strings"
	"syscall"

	"rollcage/core"

	"github.com/cactus/cobra"
	"github.com/cactus/gologit"
)

func umountCmd(args ...string) {
	cmd := exec.Command("/sbin/umount", args...)
	gologit.Debugln(append([]string{"/sbin/umount"}, args...))
	out, err := cmd.CombinedOutput()
	for _, line := range strings.Split(string(out), "\n") {
		if line != "" {
			gologit.Debugln(line)
		}
	}
	if err != nil {
		// some mounts are not present, so just fail
		// do not log exit status 1 unless debug logging
		gologit.Debugf("%s\n", err)
	}
}

func removeDash(s string) string {
	if s == "-" {
		return ""
	}
	return s
}

func stopCmdRun(cmd *cobra.Command, args []string) {
	// requires root
	if !core.IsRoot() {
		gologit.Fatalf("Must be root to stop\n")
	}

	jailUUID := core.GetJailUUIDByTagOrUUID(args[0])
	if jailUUID == "" {
		gologit.Fatalf("No jail found by '%s'\n", args[0])
	}

	jailpath := core.GetJailByTagOrUUID(jailUUID)
	if jailpath == "" {
		gologit.Fatalf("No jail found by '%s'\n", args[0])
	}

	jid := string(core.JlsMust("-j", fmt.Sprintf("ioc-%s", jailUUID), "jid"))
	if jid == "" {
		gologit.Fatalf("Jail is not running!\n")
	}

	propertyList := []string{
		"mountpoint",
		"org.freebsd.iocage:type",
		"org.freebsd.iocage:tag",
		"org.freebsd.iocage:prestop",
		"org.freebsd.iocage:exec_stop",
		"org.freebsd.iocage:poststop",
		"org.freebsd.iocage:vnet",
		"org.freebsd.iocage:ip4",
	}

	lines := core.SplitOutput(core.ZFSMust("list", "-H", "-o", strings.Join(propertyList, ","), jailpath))
	if len(lines) < 1 {
		gologit.Fatalf("No output from property fetch\n")
	}

	prop_mountpoint := removeDash(lines[0][0])
	//prop_type := removeDash(lines[0][1])
	prop_tag := removeDash(lines[0][2])
	prop_prestop := removeDash(lines[0][3])
	prop_exec_stop := removeDash(lines[0][4])
	prop_poststop := removeDash(lines[0][5])
	prop_vnet := removeDash(lines[0][6])
	prop_ip4 := removeDash(lines[0][7])

	// set a default path
	environ := []string{
		"PATH=/sbin:/bin:/usr/sbin:/usr/bin:/usr/local/sbin:/usr/local/bin",
	}

	fmt.Printf("* Stopping %s (%s)\n", jailUUID, prop_tag)
	var execErr error
	if prop_prestop != "" {
		fmt.Printf("  + Running pre-stop\n")
		preStop := core.SplitFieldsQuoteSafe(prop_prestop)
		execErr = syscall.Exec(preStop[0], preStop, environ)
		if execErr != nil {
			gologit.Printf("%s\n", execErr)
		}
	}

	fmt.Printf("  + Stopping services\n")
	jexec := []string{"/usr/sbin/jexec"}
	jexec = append(jexec, fmt.Sprintf("ioc-%s", jailUUID))
	jexec = append(jexec, core.SplitFieldsQuoteSafe(prop_exec_stop)...)
	out, err := exec.Command(jexec[0], jexec[1:]...).CombinedOutput()
	gologit.Debugln(string(out))
	if err != nil {
		gologit.Printf("%s\n", err)
	}

	if prop_vnet == "on" {
		fmt.Printf("  + Tearing down VNET\n")
		// stop VNET networking
	} else if prop_ip4 != "inherit" {
		// stop standard networking (legacy?)
		lines := core.SplitOutput(core.ZFSMust("list", "-H", "-o", "org.freebsd.iocage:ip4_addr,org.freebsd.iocage:ip6_addr", jailpath))
		for _, ip_config := range lines[0] {
			if ip_config == "none" {
				continue
			}
			for _, addr := range strings.Split(ip_config, ",") {
				item := strings.Split(addr, "|")
				gologit.Debugln("/sbin/ifconfig", item[0], item[1], "-alias")
				out, err := exec.Command("/sbin/ifconfig",
					item[0], item[1], "-alias").CombinedOutput()
				gologit.Debugln(string(out))
				if err != nil {
					gologit.Printf("%s\n", err)
				}
			}
		}
	}

	fmt.Printf("  + Removing jail process\n")
	jrexec := []string{"/usr/sbin/jail", "-r", fmt.Sprintf("ioc-%s", jailUUID)}
	out, err = exec.Command(jrexec[0], jrexec[1:]...).CombinedOutput()
	if err != nil {
		gologit.Printf("%s\n", err)
	}

	if prop_poststop != "" {
		fmt.Printf("  + Running post-stop\n")
		postStop := core.SplitFieldsQuoteSafe(prop_poststop)
		execErr := syscall.Exec(postStop[0], postStop, environ)
		if execErr != nil {
			gologit.Printf("%s\n", execErr)
		}
	}

	fmt.Printf("  + Tearing down mounts\n")
	umountCmd("-afvF", path.Join(prop_mountpoint, "fstab"))
	umountCmd(path.Join(prop_mountpoint, "root/dev/fd"))
	umountCmd(path.Join(prop_mountpoint, "root/dev"))
	umountCmd(path.Join(prop_mountpoint, "root/proc"))

	// TODO: basejail here?
	// TODO: rctl stuff here...
}

func init() {
	RootCmd.AddCommand(&cobra.Command{
		Use:   "stop UUID|TAG",
		Short: "stop jail",
		Long:  "Stop jail identified by UUID or TAG.",
		Run:   stopCmdRun,
		PreRun: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				gologit.Fatalln("Required UUID|TAG not provided")
			}
		},
	})
}
