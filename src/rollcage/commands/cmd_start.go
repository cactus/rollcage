package commands

import (
	"fmt"
	"io/ioutil"
	"os"
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

	// copy resolv conf
	err = core.CopyFile(
		"/etc/resolv.conf",
		path.Join(jail.Mountpoint, "root/etc/resolv.conf"))
	if err != nil {
		gologit.Printf("%s\n", err)
	}

	// create log file
	logfile, err := os.OpenFile(jail.GetLogPath(), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		gologit.Fatal(err)
	}
	defer logfile.Close()

	file, err := ioutil.TempFile(os.TempDir(), "rollcage.")
	defer os.Remove(file.Name())

	jailConfig := jail.JailConfig()
	gologit.Debugln(jailConfig)
	file.WriteString(jailConfig)
	file.Close()

	excmd := exec.Command(
		"/usr/sbin/jail",
		"-f", file.Name(),
		"-c", fmt.Sprintf("ioc-%s", jail.HostUUID))
	excmd.Stdout = logfile
	excmd.Stderr = logfile
	err = excmd.Run()
	if err != nil {
		gologit.Fatal(err)
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
