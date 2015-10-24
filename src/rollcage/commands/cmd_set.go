package commands

import (
	"fmt"
	"rollcage/core"
	"strings"

	"github.com/cactus/cobra"
	"github.com/cactus/gologit"
)

var CustomProperties = map[string]bool{
	"tag":              true,
	"template":         true,
	"rlimits":          true,
	"boot":             true,
	"notes":            true,
	"owner":            true,
	"priority":         true,
	"last_started":     true,
	"type":             true,
	"hostid":           true,
	"cpuset":           true,
	"jail_zfs":         true,
	"jail_zfs_dataset": true,
	"release":          true,
	"hack88":           true,
	"start":            true,
}

func ParseProps(s ...string) ([][]string, error) {
	p := [][]string{}
	for _, t := range s {
		a := strings.Split(t, "=")
		if len(a) != 2 || len(a[0]) == 0 || len(a[1]) == 0 {
			return nil, fmt.Errorf("'%s' is not a valid property", t)
		}
		p = append(p, a)
	}
	return p, nil
}

func setCmdRun(cmd *cobra.Command, args []string) {
	// requires root
	if !core.IsRoot() {
		gologit.Fatalf("Must be root to set properties\n")
	}

	if len(args) < 2 {
		gologit.Fatalln("Improper usage")
	}

	jailpath := core.GetJailByTagOrUUID(args[0])
	if jailpath == "" {
		gologit.Fatalf("No jail found by '%s'\n", args[0])
	}

	props, err := ParseProps(args[1:]...)
	if err != nil {
		gologit.Fatalln(err)
	}

	for _, prop := range props {
		prefix := ""
		if _, ok := CustomProperties[prop[0]]; ok {
			prefix = "org.freebsd.iocage:"
		}
		zfsArgs := []string{
			"set",
			fmt.Sprintf("%s%s=%s", prefix, prop[0], prop[1]),
			jailpath,
		}
		gologit.Debugln(zfsArgs)
		core.ZFSMust(zfsArgs...)
	}
}

func init() {
	RootCmd.AddCommand(&cobra.Command{
		Use:   "set UUID|TAG property=value [property=value ...]",
		Short: "Set a property to a given value",
		Run:   setCmdRun,
	})
}
