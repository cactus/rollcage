package commands

import (
	"fmt"
	"io/ioutil"
	"os"
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
	defer f.Close()

	fmt.Printf("* Stopping %s (%s)\n", jail.HostUUID, jail.Tag)
	fmt.Printf("  + Removing jail process\n")

	file, err := ioutil.TempFile(os.TempDir(), "rollcage.")
	defer os.Remove(file.Name())

	jailConfig := jail.JailConfig()
	gologit.Debugln(jailConfig)
	file.WriteString(jailConfig)
	file.Close()

	excmd := exec.Command(
		"/usr/sbin/jail",
		"-f", file.Name(),
		"-r", fmt.Sprintf("ioc-%s", jail.HostUUID))
	excmd.Stdout = f
	excmd.Stderr = f
	err = excmd.Run()
	if err != nil {
		gologit.Fatal(err)
	}

	// mostly for safety...
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
