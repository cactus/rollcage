package commands

import (
	"fmt"
	"path"
	"strings"
	"syscall"

	"rollcage/core"

	"github.com/cactus/cobra"
	"github.com/cactus/gologit"
)

func umountCmd(environ []string, args ...string) {
	umountcmd := []string{"/sbin/umount"}
	umountcmd = append(umountcmd, args...)
	execErr := syscall.Exec(umountcmd[0], umountcmd, environ)
	if execErr != nil {
		gologit.Printf("%s\n", execErr)
	}
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
		"org.freebsd.iocage:type",
		"org.freebsd.iocage:mountpoint",
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

	//prop_type := lines[0][0]
	prop_mountpoint := lines[0][1]
	prop_tag := lines[0][2]
	prop_prestop := lines[0][3]
	prop_exec_stop := lines[0][4]
	prop_poststop := lines[0][5]
	prop_vnet := lines[0][6]
	prop_ip4 := lines[0][7]

	fmt.Printf("* Stopping %s (%s)", jailUUID, prop_tag)

	// set a default path
	environ := []string{
		"PATH=/sbin:/bin:/usr/sbin:/usr/bin:/usr/local/sbin:/usr/local/bin",
	}

	if prop_prestop != "" {
		fmt.Printf("  + Running pre-stop")
		preStop := core.SplitFieldsQuoteSafe(prop_prestop)
		execErr := syscall.Exec(preStop[0], preStop, environ)
		if execErr != nil {
			gologit.Printf("%s\n", execErr)
		}
	}

	fmt.Printf("  + Stopping services")
	jexec := []string{"/usr/sbin/jexec"}
	jexec = append(jexec, fmt.Sprintf("ioc-%s", jailUUID))
	jexec = append(jexec, core.SplitFieldsQuoteSafe(prop_exec_stop)...)
	execErr := syscall.Exec(jexec[0], jexec, environ)
	if execErr != nil {
		gologit.Printf("%s\n", execErr)
	}

	if prop_vnet != "" {
		fmt.Printf("  + Tearing down VNET")
		// stop networking
	} else {
		if prop_ip4 != "inherit" {
			// stop legacy net
		}
	}

	fmt.Printf("  + Removing jail process")
	jrexec := []string{"/usr/sbin/jail", "-r", fmt.Sprintf("ioc-%s", jailUUID)}
	execErr = syscall.Exec(jrexec[0], jrexec, environ)
	if execErr != nil {
		gologit.Printf("%s\n", execErr)
	}

	if prop_poststop != "" {
		fmt.Printf("  + Running post-stop")
		postStop := core.SplitFieldsQuoteSafe(prop_poststop)
		execErr := syscall.Exec(postStop[0], postStop, environ)
		if execErr != nil {
			gologit.Printf("%s\n", execErr)
		}
	}

	fmt.Printf("  + Tearing down mounts")
	umountCmd(environ, "-afvF", path.Join(prop_mountpoint, "fstab"))
	umountCmd(environ, path.Join(prop_mountpoint, "root/dev/fd"))
	umountCmd(environ, path.Join(prop_mountpoint, "root/dev"))
	umountCmd(environ, path.Join(prop_mountpoint, "root/proc"))

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
