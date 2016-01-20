package commands

import (
	"fmt"

	"rollcage/core"

	"github.com/cactus/cobra"
)

func releaseListCmdRun(cmd *cobra.Command, args []string) {
	releases := core.GetAllReleases()

	wf := core.NewOutputWriter([]string{"name"}, MachineOutput)
	for _, release := range releases {
		/*
			        if strings.HasPrefix(line, "-") {
						continue
					}
		*/
		fmt.Fprintf(wf, "%s\n", release.Name)
	}
	wf.Flush()
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

	RootCmd.AddCommand(ReleaseCmd)
}
