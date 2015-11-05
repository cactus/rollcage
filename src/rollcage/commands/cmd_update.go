package commands

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"rollcage/core"

	"github.com/cactus/cobra"
	"github.com/cactus/gologit"
)

func updateCmdRun(cmd *cobra.Command, args []string) {
	// requires root
	if !core.IsRoot() {
		gologit.Fatalf("Must be root to snapremove\n")
	}

	jailpath := core.GetJailByTagOrUUID(args[0])
	if jailpath == "" {
		gologit.Fatalf("No jail found by '%s'\n", args[0])
	}

	zfsArgs := []string{
		"get", "-Ho", "value", "org.freebsd.iocage:release,mountpoint",
		jailpath}
	out := strings.Split(
		strings.TrimSpace(string(core.ZFSMust(zfsArgs...))), "\n")
	release := out[0]
	mountpoint := out[1]

	resolvconf := path.Join(mountpoint, "root/etc/resolv.conf")
	if _, err := os.Stat(resolvconf); os.IsNotExist(err) {
		data, err := ioutil.ReadFile("/etc/resolv.conf")
		if err != nil {
			gologit.Fatalln("/etc/resolv.conf not present or not readable")
		}

		err = ioutil.WriteFile(resolvconf, data, 0755)
		if err != nil {
			gologit.Fatalf("Could not copy contents to '%s'\n", resolvconf)
		}
	}

	fmt.Println("* Creating back out snapshot")
	snappath := fmt.Sprintf(
		"%s/root@%s",
		jailpath,
		fmt.Sprintf(
			"ioc-update-%s",
			time.Now().Format("2006-01-02_15:04:05")))
	core.ZFSMust("snapshot", snappath)

	devroot := path.Join(mountpoint, "root/dev")
	ecmd := exec.Command("/sbin/mount", "-t", "devfs", "devfs", devroot)
	gologit.Debugln(ecmd.Args)
	eout, err := ecmd.CombinedOutput()
	if err != nil {
		gologit.Fatalf("Error mounting devfs: %s\n", err)
	}
	gologit.Debugln(string(eout))

	defer func() {
		ecmd := exec.Command("/sbin/umount", devroot)
		gologit.Debugln(ecmd.Args)
		err := ecmd.Run()
		if err != nil {
			gologit.Fatalf("Error unmounting devfs: %s\n", err)
		}
	}()

	fmt.Println("* Updating jail...")
	root := path.Join(mountpoint, "root")
	ecmd = exec.Command("/usr/sbin/chroot", root,
		"/usr/sbin/freebsd-update", "--not-running-from-cron",
		"fetch", "install")
	ecmd.Env = []string{
		"PATH=/sbin:/bin:/usr/sbin:/usr/bin:/usr/local/sbin:/usr/local/bin",
		fmt.Sprintf("UNAME_r=%s", release),
		"PAGER=/bin/cat",
	}
	gologit.Debugln(ecmd.Args)
	ecmd.Stdout = os.Stdout
	ecmd.Stderr = os.Stderr
	ecmd.Run()

	fmt.Println("* update finished")
	fmt.Println("  Once verified, don't forget to remove the snapshot!")
}

func init() {
	cmd := &cobra.Command{
		Use:   "update UUID|TAG",
		Short: "update a jail to the latest patchlevel",
		Run:   updateCmdRun,
	}

	RootCmd.AddCommand(cmd)
}
