package commands

import (
	"strings"

	"rollcage/core"

	"github.com/cactus/cobra"
	"github.com/cactus/gologit"
)

func getCmdRun(cmd *cobra.Command, args []string) {
	headers := getOutputColsFlag.GetCols()
	if len(headers) == 0 {
		headers = getOutputColsFlag.GetValidCols()
	}

	var allTags bool
	desiredTags := make(map[string]bool)
	if len(args) == 0 || args[0] == "all" {
		allTags = true
	} else {
		allTags = false
		for _, prop := range strings.Split(args[0], ",") {
			desiredTags[prop] = true
		}
	}

	var jails []*core.JailMeta
	if len(args) < 2 {
		jails = core.GetAllJails()
	} else {
		jail, err := core.FindJail(args[1])
		if err != nil {
			gologit.Fatalf("No jail found by '%s'\n", args[1])
		}
		jails = append(jails, jail)
	}

	twf := core.NewTemplateOutputWriter(headers, MachineOutput)
	for _, jail := range jails {
		zfsArgs := []string{"get", "-H"}
		if ParsableValues {
			zfsArgs = append(zfsArgs, "-p")
		}
		zfsArgs = append(zfsArgs, "all", jail.Path)
		out := core.ZFSMust(zfsArgs...)

		properties := make([][]string, 0)
		for _, line := range strings.Split(string(out), "\n") {
			if line == "" {
				continue
			}
			cols := strings.Split(line, "\t")
			property := cols[1]
			if strings.HasPrefix(property, "org.freebsd.iocage:") {
				property = strings.Split(property, ":")[1]
			}
			if property == "tag" {
				continue
			}
			if property == "host_hostuuid" {
				continue
			}
			if allTags || desiredTags[property] {
				properties = append(properties, []string{property, cols[2]})
			}
		}
		for _, prop := range properties {
			twf.WriteTemplate(&struct {
				Uuid     string
				Tag      string
				Property string
				Value    string
			}{
				Uuid:     jail.HostUUID,
				Tag:      jail.Tag,
				Property: prop[0],
				Value:    prop[1],
			})
		}
	}
	twf.Flush()
}

// sorted list of output headers (lowercase for matching)
var getOutputColsFlag = core.NewOutputCols(
	[]string{"uuid", "tag", "property", "value"})

func init() {
	cmd := &cobra.Command{
		Use:   "get all|property[,property]... [UUID|TAG]",
		Short: "get list of properties",
		Long:  "Get named property or if \"all\" keyword is specified dump all properties known to iocage.",
		Run:   getCmdRun,
	}

	cmd.Flags().VarP(
		getOutputColsFlag, "output-columns", "o",
		"output columns")
	cmd.Flags().BoolVarP(
		&ParsableValues, "parsable-values", "p",
		false, "output parsable (exact) values")

	RootCmd.AddCommand(cmd)
}
