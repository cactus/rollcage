package commands

import (
	"fmt"
	"os/exec"
	"path"
	"time"

	"rollcage/core"

	"github.com/cactus/cobra"
	"github.com/cactus/gologit"
)

func startCmdRun(cmd *cobra.Command, args []string) {
	// requires root
	if !core.IsRoot() {
		gologit.Fatalf("Must be root to stop\n")
	}

	jail, err := core.FindJail(args[0])
	if err != nil {
		gologit.Fatalf("No jail found by '%s'\n", args[0])
	}

	if jail.IsRunning() {
		gologit.Fatalf("Jail is already running!\n")
	}

	props := jail.GetProperties()

	// set a default path
	environ := []string{
		"PATH=/sbin:/bin:/usr/sbin:/usr/bin:/usr/local/sbin:/usr/local/bin",
	}

	fmt.Printf("* Starting %s (%s)\n", jail.HostUUID, jail.Tag)
	// mount procfs
	if props.GetIOC("mount_procfs") == "1" {
		fmt.Printf("  + mounting procfs\n")
		procpath := path.Join(jail.Mountpoint, "root/proc")
		excmd := exec.Command("/sbin/mount", "-t", "procfs", "proc", procpath)
		excmd.Env = environ
		err := excmd.Run()
		if err != nil {
			gologit.Printf("%s\n", err)
		}
	}

	// prepare jail zfs dataset if enabled
	if props.GetIOC("jail_zfs") == "on" {
		fmt.Printf("  + jailing zfs dataset\n")
		setprops := core.ZFSProperties{
			"org.freebsd.iocage:allow_mount":     "1",
			"org.freebsd.iocage:allow_mount_zfs": "1",
			"org.freebsd.iocage:enforce_statfs":  "1",
		}
		jail.SetProperties(setprops)
		core.ZFSMust(
			fmt.Errorf("Error setting property"),
			"set", "jailed=on",
			path.Join(core.GetZFSRootPath(), props.GetIOC("jail_zfs_dataset")))
	}

	// bring up networking (vnet or legacy)
	ip4_addr_propline := "ip4.addr="
	ip6_addr_propline := "ip6.addr="
	if props.GetIOC("vnet") == "on" {
		fmt.Printf("  + Configuring VNET\n")
		// start VNET networking
	} else {
		// start standard networking (legacy?)
		if props.GetIOC("ip4_addr") != "none" {
			ip4_addr_propline += props.GetIOC("ip4_addr")
		}
		if props.GetIOC("ip6_addr") != "none" {
			ip6_addr_propline += props.GetIOC("ip6_addr")
		}
	}

	// get log dir

	logdir := core.ZFSMust(
		fmt.Errorf("Error setting property"),
		"get", "-H", "-o", "value", "mountpoint", core.GetZFSRootPath())
	logdir = path.Join(logdir, "log")

	logpath := path.Join(logdir, fmt.Sprintf("%s-console.log", jail.HostUUID))

	// start jail
	jailexec := []string{
		"/usr/sbin/jail", "-c",
		ip4_addr_propline,
		fmt.Sprintf("ip4.saddrsel=%s", props.GetIOC("ip4_saddrsel")),
		fmt.Sprintf("ip4=%s", props.GetIOC("ip4")),
		ip6_addr_propline,
		fmt.Sprintf("ip6.saddrsel=%s", props.GetIOC("ip6_saddrsel")),
		fmt.Sprintf("ip6=%s", props.GetIOC("ip6")),
		fmt.Sprintf("name=ioc-%s", jail.HostUUID),
		fmt.Sprintf("host.hostname=%s", props.GetIOC("hostname")),
		fmt.Sprintf("host.hostuuid=%s", props.GetIOC("host_hostuuid")),
		fmt.Sprintf("path=%s", path.Join(jail.Mountpoint, "root")),
		fmt.Sprintf("securelevel=%s", props.GetIOC("securelevel")),
		fmt.Sprintf("devfs_ruleset=%s", props.GetIOC("devfs_ruleset")),
		fmt.Sprintf("enforce_statfs=%s", props.GetIOC("enforce_statfs")),
		fmt.Sprintf("children.max=%s", props.GetIOC("children_max")),
		fmt.Sprintf("allow.set_hostname=%s", props.GetIOC("allow_set_hostname")),
		fmt.Sprintf("allow.sysvipc=%s", props.GetIOC("allow_sysvipc")),
		fmt.Sprintf("allow.chflags=%s", props.GetIOC("allow_chflags")),
		fmt.Sprintf("allow.mount=%s", props.GetIOC("allow_mount")),
		fmt.Sprintf("allow.mount.devfs=%s", props.GetIOC("allow_mount_devfs")),
		fmt.Sprintf("allow.mount.nullfs=%s", props.GetIOC("allow_mount_nullfs")),
		fmt.Sprintf("allow.mount.procfs=%s", props.GetIOC("allow_mount_procfs")),
		fmt.Sprintf("allow.mount.tmpfs=%s", props.GetIOC("allow_mount_tmpfs")),
		fmt.Sprintf("allow.mount.zfs=%s", props.GetIOC("allow_mount_zfs")),
		fmt.Sprintf("mount.fdescfs=%s", props.GetIOC("mount_fdescfs")),
		fmt.Sprintf("allow.quotas=%s", props.GetIOC("allow_quotas")),
		fmt.Sprintf("allow.socket_af=%s", props.GetIOC("allow_socket_af")),
		fmt.Sprintf("exec.prestart=%s", props.GetIOC("prestart")),
		fmt.Sprintf("exec.poststart=%s", props.GetIOC("poststart")),
		fmt.Sprintf("exec.prestop=%s", props.GetIOC("prestop")),
		fmt.Sprintf("exec.stop=%s", props.GetIOC("exec_stop")),
		fmt.Sprintf("exec.clean=%s", props.GetIOC("exec_clean")),
		fmt.Sprintf("exec.timeout=%s", props.GetIOC("exec_timeout")),
		fmt.Sprintf("exec.fib=%s", props.GetIOC("exec_fib")),
		fmt.Sprintf("stop.timeout=%s", props.GetIOC("stop_timeout")),
		fmt.Sprintf("mount.fstab=%s", path.Join(jail.Mountpoint, "fstab")),
		fmt.Sprintf("mount.devfs=%s", props.GetIOC("mount_devfs")),
		fmt.Sprintf("exec.consolelog=%s", logpath),
		"allow.dying",
		"persist",
	}
	gologit.Debugln(jailexec)
	out, err := exec.Command(jailexec[0], jailexec[1:]...).CombinedOutput()
	gologit.Debugln(string(out))
	if err != nil {
		gologit.Printf("%s\n", err)
	}

	// rctl_limits?
	// cpuset?

	// jail zfs
	if props.GetIOC("jail_zfs") == "on" {
		core.ZFSMust(
			fmt.Errorf("Error setting property"),
			"jail", fmt.Sprintf("ioc-%s", jail.HostUUID),
			path.Join(core.GetZFSRootPath(), props.GetIOC("jail_zfs_dataset")))
		out, err := exec.Command(
			"/usr/sbin/jexec",
			fmt.Sprintf("ioc-%s", jail.HostUUID),
			"zfs", "mount", "-a").CombinedOutput()
		gologit.Debugln(string(out))
		if err != nil {
			gologit.Printf("%s\n", err)
		}
	}

	// copy resolv conf
	err = core.CopyFile(
		"/etc/resolv.conf",
		path.Join(jail.Mountpoint, "root/etc/resolv.conf"))
	if err != nil {
		gologit.Printf("%s\n", err)
	}

	// start services
	fmt.Printf("  + Starting services\n")
	jexec := []string{}
	if props.GetIOC("exec_fib") != "0" {
		jexec = append(jexec, "/usr/sbin/setfib", props.GetIOC("exec_fib"))
	}

	jexec = append(jexec, "/usr/sbin/jexec")
	jexec = append(jexec, fmt.Sprintf("ioc-%s", jail.HostUUID))
	jexec = append(jexec, core.SplitFieldsQuoteSafe(props.GetIOC("exec_start"))...)
	out, err = exec.Command(jexec[0], jexec[1:]...).CombinedOutput()
	gologit.Debugln(string(out))
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
		Use:   "start UUID|TAG",
		Short: "start jail",
		Long:  "Start jail identified by UUID or TAG.",
		Run:   startCmdRun,
		PreRun: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				gologit.Fatalln("Required UUID|TAG not provided")
			}
		},
	})
}
