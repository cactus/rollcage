package commands

import (
	"fmt"
	"net/url"
	"os"
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

	release, err := core.FindRelease(args[0])
	if err == nil {
		gologit.Fatalf("Release '%s' already exists!\n", release.Name)
	}

	// get some meta
	release, err = core.CreateRelease(releaseName)
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
	fmt.Printf("Fetching sets\n")
	for _, setname := range strings.Split(fetchSets, " ") {
		ux := *u
		ux.Path = path.Join(ux.Path, release.Name, setname)
		destPth := path.Join(release.Mountpoint, "sets", setname)
		if _, err := os.Stat(destPth); !os.IsNotExist(err) {
			fmt.Printf("'%s' already present. Skipping download.\n", setname)
			continue
		}
		err := core.FetchHTTPFile(ux.String(), destPth, true)
		if err != nil {
			gologit.Fatalf("Failed to fetch: %s\n", ux.String())
		}
	}
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
		&mirrorHost, "mirror-host", "", "ftp.freebsd.org", "set mirror hostname")
	fetchCommand.Flags().StringVarP(
		&mirrorDir, "mirror-dir", "", "/pub/FreeBSD/releases/amd64/amd64/", "set mirror hostname")
	fetchCommand.Flags().StringVarP(
		&fetchSets, "sets", "s", "base.txz doc.txz lib32.txz src.txz",
		"sets to fetch for a release")

	ReleaseCmd.AddCommand(fetchCommand)
	RootCmd.AddCommand(ReleaseCmd)
}
