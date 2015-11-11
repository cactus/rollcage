package commands

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/cactus/cobra"
)

var Version = "no-version"
var LicenseText = `
This software is available under the MIT License.
    https://github.com/cactus/rollcage

Portions of this software utilize third party libraries:
*   https://github.com/cactus/cobra
    Forked from: https://github.com/spf13/cobra (Apache 2.0 License)
*   https://github.com/cactus/gologit (MIT license)
*   https://github.com/spf13/pflag (BSD license)
*   https://github.com/go-gcfg/gcfg/tree/v1 (BSD license)
`
var showLicense bool

func init() {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("rollcage %s (%s,%s-%s)\n\n", Version,
				runtime.Version(), runtime.Compiler, runtime.GOARCH)
			if showLicense {
				fmt.Printf("%s\n", strings.TrimSpace(LicenseText))
			}
		},
	}
	cmd.Flags().BoolVarP(
		&showLicense, "license", "l", false,
		"output information about licenses and dependencies")
	RootCmd.AddCommand(cmd)
}
