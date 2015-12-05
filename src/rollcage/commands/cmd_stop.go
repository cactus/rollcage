package commands

import (
	"fmt"
	"os/exec"
	"path"
	"strings"

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

	jail, err := core.FindJail(args[0])
	if err != nil {
		gologit.Fatalf("No jail found by '%s'\n", args[0])
	}

	if !jail.IsRunning() {
		gologit.Fatalf("Jail is not running!\n")
	}

	props := jail.GetProperties()

	fmt.Printf("* Stopping %s (%s)\n", jail.HostUUID, jail.Tag)
	fmt.Printf("  + Removing jail process\n")
	core.CmdMust(
		fmt.Errorf("Error stopping jail!"),
		"/usr/sbin/jail", "-r", fmt.Sprintf("ioc-%s", jail.HostUUID))

	if props.GetIOC("vnet") == "on" {
		fmt.Printf("  + Tearing down VNET\n")
		// stop VNET networking
	} else if props.GetIOC("ip4") != "inherit" {
		// stop standard networking (legacy?)
		lines := core.SplitOutput(core.ZFSMust(
			fmt.Errorf("Error listing jails"),
			"list", "-H", "-o", "org.freebsd.iocage:ip4_addr,org.freebsd.iocage:ip6_addr", jail.Path))
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

	fmt.Printf("  + Tearing down mounts\n")
	umountCmd("-afvF", path.Join(jail.Mountpoint, "fstab"))
	umountCmd(path.Join(jail.Mountpoint, "root/dev/fd"))
	umountCmd(path.Join(jail.Mountpoint, "root/dev"))
	umountCmd(path.Join(jail.Mountpoint, "root/proc"))

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
