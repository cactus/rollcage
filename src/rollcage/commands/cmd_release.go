package commands

import (
	"fmt"
	"os"
	"strings"

	"rollcage/core"

	"github.com/cactus/cobra"
	"github.com/cactus/gologit"
)

func releaseListCmdRun(cmd *cobra.Command, args []string) {
	releases := core.GetAllReleases()

	wf := core.NewOutputWriter([]string{"name", "patchlevel"}, MachineOutput)
	for _, release := range releases {
		fmt.Fprintf(wf, "%s\t%s\n", release.Name, release.Patchlevel)
	}
	wf.Flush()
}

func releaseFetchCmdRun(cmd *cobra.Command, args []string) {
	// requires root
	if !core.IsRoot() {
		gologit.Fatalf("Must be root to fetch\n")
	}

	/*
		release, err := core.FindRelease(args[0])
		if err == nil {
			gologit.Fatalf("Release '%s' already exists!\n", args[0])
		}
	*/

	// create
	// fetch
	// update?
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

	RootCmd.AddCommand(ReleaseCmd)
}
