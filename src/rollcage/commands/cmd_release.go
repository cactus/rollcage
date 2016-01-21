package commands

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strings"

	"rollcage/core"

	"github.com/cactus/cobra"
	"github.com/cactus/gologit"
)

var (
	fetchSets  string
	mirrorHost string
	mirrorDir  string
)

func releaseListCmdRun(cmd *cobra.Command, args []string) {
	releases := core.GetAllReleases()

	wf := core.NewOutputWriter([]string{"name", "patchlevel"}, MachineOutput)
	for _, release := range releases {
		fmt.Fprintf(wf, "%s\t%s\n", release.Name, release.Patchlevel)
	}
	wf.Flush()
}

func releaseDestroyCmdRun(cmd *cobra.Command, args []string) {
	// requires root
	if !core.IsRoot() {
		gologit.Fatalf("Must be root to destroy\n")
	}

	release, err := core.FindRelease(args[0])
	if err != nil {
		gologit.Fatalf("Release '%s' not found!\n", args[0])
	}

	fmt.Printf("Ready to remove release: %s\n", release.Name)
	if !force {
		fmt.Print("Are you sure [yN]? ")
		var response string
		_, err = fmt.Scanln(&response)
		if err != nil {
			if err.Error() == "unexpected newline" {
				os.Exit(0)
			}
			gologit.Fatalf("%s", err)
		}

		response = strings.ToLower(strings.TrimSpace(response))
		if len(response) != 1 || response[0] != 'y' {
			return
		}
	}

	fmt.Printf("Destroying: %s\n", release.Name)
	core.ZFSMust(
		fmt.Errorf("Error destroying release"),
		"destroy", "-fr", release.Path)
	os.RemoveAll(release.Mountpoint)
}

func releaseFetchCmdRun(cmd *cobra.Command, args []string) {
	// requires root
	if !core.IsRoot() {
		gologit.Fatalf("Must be root to fetch\n")
	}

	// find/verify release name
	releaseName := strings.ToUpper(args[0])

	found := false
	for _, release := range core.SupportedReleases {
		if releaseName == release {
			found = true
			break
		}
	}

	if !found {
		gologit.Fatalf("Release '%s' is not currently supported\n", releaseName)
	}

	// get some meta
	release, err := core.CreateRelease(releaseName)
	if err != nil {
		gologit.Fatalf("Couldn't create release '%s'\n", releaseName)
	}

	// fetch
	if !strings.HasPrefix(mirrorHost, "http://") {
		mirrorHost = "http://" + mirrorHost
	}
	u, err := url.Parse(mirrorHost)
	if err != nil {
		gologit.Fatalf("error parsing internal sets fetch url\n")
	}
	u.Path = mirrorDir
	fmt.Printf("Fetching sets:\n")
	tarset := []string{}
	for _, setname := range strings.Split(fetchSets, " ") {
		ux := *u
		ux.Path = path.Join(ux.Path, release.Name, setname)
		destPth := path.Join(release.Mountpoint, "sets", setname)
		tarset = append(tarset, destPth)
		if _, err := os.Stat(destPth); !os.IsNotExist(err) {
			fmt.Printf("'%s' already present -- skipping download\n", setname)
			continue
		}
		err := core.FetchHTTPFile(ux.String(), destPth, true)
		if err != nil {
			gologit.Fatalf("Failed to fetch: %s\n", ux.String())
		}
	}

	fmt.Printf("Extracting sets:\n")
	for _, pth := range tarset {
		basepth := path.Base(pth)
		fmt.Printf("* %s\n", basepth)
		excmd := exec.Command(
			"tar", "-C", path.Join(release.Mountpoint, "root"),
			"-xf", pth)
		excmd.Stdout = os.Stdout
		excmd.Stderr = os.Stdout
		err := excmd.Run()
		if err != nil {
			gologit.Debugf("Error: %s\n", err)
			gologit.Fatalf("Failed to extract: %s\n", basepth)
		}
	}
	err = os.MkdirAll(path.Join(release.Mountpoint, "root", "usr/home"), 0755)
	if err != nil {
		gologit.Fatalf("Failed to make: %s\n", "/usr/home")
	}
	err = os.MkdirAll(path.Join(release.Mountpoint, "root", "usr/ports"), 0755)
	if err != nil {
		gologit.Fatalf("Failed to make: %s\n", "/usr/ports")
	}
}

func releaseUpdateCmdRun(cmd *cobra.Command, args []string) {
	// requires root
	if !core.IsRoot() {
		gologit.Fatalf("Must be root to update\n")
	}

	release, err := core.FindRelease(args[0])
	if err != nil {
		gologit.Fatalf("Release '%s' not found!\n", args[0])
	}

	mountpoint := release.Mountpoint
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

	fmt.Println("* Updating release...")
	root := path.Join(mountpoint, "root")

	exargs := []string{root, "/usr/sbin/freebsd-update"}
	if release.Name != "9.3-RELEASE" && release.Name != "10.1-RELEASE" {
		exargs = append(exargs, "--not-running-from-cron")
	}
	exargs = append(exargs, "fetch", "install")
	ecmd = exec.Command("/usr/sbin/chroot", exargs...)
	ecmd.Env = []string{
		"PATH=/sbin:/bin:/usr/sbin:/usr/bin:/usr/local/sbin:/usr/local/bin",
		fmt.Sprintf("UNAME_r=%s", release.Name),
		"PAGER=/bin/cat",
	}
	gologit.Debugln(ecmd.Args)
	ecmd.Stdout = os.Stdout
	ecmd.Stderr = os.Stderr
	ecmd.Stdin = os.Stdin
	ecmd.Run()
	fmt.Println("* update finished")
}

func init() {
	ReleaseCmd := &cobra.Command{
		Use:   "release",
		Short: "Operations for listing and fetching releases",
	}

	ReleaseCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all releases",
		Run:   releaseListCmdRun,
	})

	ReleaseCmd.AddCommand(&cobra.Command{
		Use:   "destroy RELEASE",
		Short: "Remove a release",
		Run:   releaseDestroyCmdRun,
		PreRun: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				gologit.Fatalln("Required RELEASE not provided")
			}
		},
	})

	fetchCommand := &cobra.Command{
		Use:   "fetch RELEASE",
		Short: "Fetch/add a release",
		Run:   releaseFetchCmdRun,
		PreRun: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				gologit.Fatalln("Required RELEASE not provided")
			}
		},
	}

	fetchCommand.Flags().StringVarP(
		&mirrorHost, "mirror-host", "", "ftp.freebsd.org",
		"set mirror hostname")
	fetchCommand.Flags().StringVarP(
		&mirrorDir, "mirror-dir", "", "/pub/FreeBSD/releases/amd64/amd64/",
		"set mirror hostname")
	fetchCommand.Flags().StringVarP(
		&fetchSets, "sets", "s", "base.txz doc.txz lib32.txz src.txz",
		"sets to fetch for a release")

	ReleaseCmd.AddCommand(fetchCommand)

	ReleaseCmd.AddCommand(&cobra.Command{
		Use:   "update RELEASE",
		Short: "Update a release to most recent patchset",
		Run:   releaseUpdateCmdRun,
		PreRun: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				gologit.Fatalln("Required RELEASE not provided")
			}
		},
	})

	RootCmd.AddCommand(ReleaseCmd)
}
