package commands

import (
	"fmt"
	"runtime"

	"github.com/cactus/cobra"
)

var Version = "no-version"

func init() {
	RootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("rollcage %s (%s,%s-%s)\n", Version,
				runtime.Version(), runtime.Compiler, runtime.GOARCH)
		},
	})
}
