// Copyright © 2013 Steve Francia <spf@spf13.com>.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//Package cobra is a commander providing a simple interface to create powerful
//modern CLI interfaces.
//In addition to providing an interface, Cobra simultaneously provides a
//controller to organize your application code.
package cobra

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	flag "github.com/spf13/pflag"
)

// Command is just that, a command for your application.
// eg.  'go run' ... 'run' is the command. Cobra requires
// you to define the usage and description as part of your command
// definition to ensure usability.
type Command struct {
	// Name is the command name, usually the executable's name.
	name string
	// The one-line usage message.
	Use string
	// An array of aliases that can be used instead of the first word in Use.
	Aliases []string
	// The short description shown in the 'help' output.
	Short string
	// The long message shown in the 'help <this-command>' output.
	Long string
	// Examples of how to use the command
	Example string

	// The *Run functions are executed in the following order:
	//   * PersistentPreRun()
	//   * PreRun()
	//   * Run()
	//   * PostRun()
	//   * PersistentPostRun()
	// All functions get the same args, the arguments after the command name
	// PersistentPreRun: children of this command will inherit and execute
	PersistentPreRun func(cmd *Command, args []string)
	// PreRun: children of this command will not inherit.
	PreRun func(cmd *Command, args []string)
	// Run: Typically the actual work function. Most commands will only implement this
	Run func(cmd *Command, args []string)
	// PostRun: run after the Run command.
	PostRun func(cmd *Command, args []string)
	// PersistentPostRun: children of this command will inherit and execute after PostRun
	PersistentPostRun func(cmd *Command, args []string)

	// Full set of flags
	flags *flag.FlagSet
	// Set of flags childrens of this command will inherit
	pflags *flag.FlagSet
	// Flags that are declared specifically by this command (not inherited).
	lflags *flag.FlagSet

	// Commands is the list of commands supported by this program.
	commands []*Command
	// Parent Command for this command
	parent *Command
	// max lengths of commands' string lengths for use in padding
	commandsMaxUseLen         int
	commandsMaxCommandPathLen int
	commandsMaxNameLen        int

	flagErrorBuf *bytes.Buffer

	args          []string                 // actual args parsed from flags
	output        *io.Writer               // nil means stderr; use Out() method instead
	usageTemplate string                   // Can be defined by Application
	usageFunc     func(*Command) error     // Usage can be defined by application
	helpTemplate  string                   // Can be defined by Application
	helpFunc      func(*Command, []string) // Help can be defined by application
	helpCommand   *Command                 // The help command
	// The global normalization function that we can use on every pFlag set and children commands
	globNormFunc func(f *flag.FlagSet, name string) flag.NormalizedName
}

// os.Args[1:] by default, if desired, can be overridden
// particularly useful when testing.
func (c *Command) SetArgs(a []string) {
	c.args = a
}

func (c *Command) GetOutOr(def io.Writer) io.Writer {
	if c.output != nil {
		return *c.output
	}

	if c.HasParent() {
		return c.parent.Out()
	} else {
		return def
	}
}

func (c *Command) Out() io.Writer {
	return c.GetOutOr(os.Stderr)
}

// SetOutput sets the destination for usage and error messages.
// If output is nil, os.Stderr is used.
func (c *Command) SetOutput(output io.Writer) {
	c.output = &output
}

// Usage can be defined by application
func (c *Command) SetUsageFunc(f func(*Command) error) {
	c.usageFunc = f
}

// Can be defined by Application
func (c *Command) SetUsageTemplate(s string) {
	c.usageTemplate = s
}

// Can be defined by Application
func (c *Command) SetHelpFunc(f func(*Command, []string)) {
	c.helpFunc = f
}

func (c *Command) SetHelpCommand(cmd *Command) {
	c.helpCommand = cmd
}

// Can be defined by Application
func (c *Command) SetHelpTemplate(s string) {
	c.helpTemplate = s
}

// SetGlobalNormalizationFunc sets a normalization function to all flag sets and also to child commands.
// The user should not have a cyclic dependency on commands.
func (c *Command) SetGlobalNormalizationFunc(n func(f *flag.FlagSet, name string) flag.NormalizedName) {
	c.Flags().SetNormalizeFunc(n)
	c.PersistentFlags().SetNormalizeFunc(n)
	c.globNormFunc = n

	for _, command := range c.commands {
		command.SetGlobalNormalizationFunc(n)
	}
}

func (c *Command) UsageFunc() (f func(*Command) error) {
	if c.usageFunc != nil {
		return c.usageFunc
	}

	if c.HasParent() {
		return c.parent.UsageFunc()
	} else {
		return func(c *Command) error {
			err := tmpl(c.Out(), c.UsageTemplate(), c)
			if err != nil {
				fmt.Print(err)
			}
			return err
		}
	}
}

// HelpFunc returns either the function set by SetHelpFunc for this command
// or a parent, or it returns a function which calls c.Help()
func (c *Command) HelpFunc() func(*Command, []string) {
	cmd := c
	for cmd != nil {
		if cmd.helpFunc != nil {
			return cmd.helpFunc
		}
		cmd = cmd.parent
	}
	return func(*Command, []string) {
		err := c.Help()
		if err != nil {
			c.Println(err)
		}
	}
}

var minUsagePadding int = 25

func (c *Command) UsagePadding() int {
	if c.parent == nil || minUsagePadding > c.parent.commandsMaxUseLen {
		return minUsagePadding
	} else {
		return c.parent.commandsMaxUseLen
	}
}

var minCommandPathPadding int = 11

//
func (c *Command) CommandPathPadding() int {
	if c.parent == nil || minCommandPathPadding > c.parent.commandsMaxCommandPathLen {
		return minCommandPathPadding
	} else {
		return c.parent.commandsMaxCommandPathLen
	}
}

var minNamePadding int = 11

func (c *Command) NamePadding() int {
	if c.parent == nil || minNamePadding > c.parent.commandsMaxNameLen {
		return minNamePadding
	} else {
		return c.parent.commandsMaxNameLen
	}
}

func (c *Command) UsageTemplate() string {
	if c.usageTemplate != "" {
		return c.usageTemplate
	}

	if c.HasParent() {
		return c.parent.UsageTemplate()
	} else {
		return defaultUsageTemplate
	}
}

func (c *Command) HelpTemplate() string {
	if c.helpTemplate != "" {
		return c.helpTemplate
	}

	if c.HasParent() {
		return c.parent.HelpTemplate()
	} else {
		return defaultHelpTemplate
	}
}

// Really only used when casting a command to a commander
func (c *Command) resetChildrensParents() {
	for _, x := range c.commands {
		x.parent = c
	}
}

// Test if the named flag is a boolean flag.
func isBooleanFlag(name string, f *flag.FlagSet) bool {
	flag := f.Lookup(name)
	if flag == nil {
		return false
	}
	return flag.Value.Type() == "bool"
}

// Test if the named flag is a boolean flag.
func isBooleanShortFlag(name string, f *flag.FlagSet) bool {
	result := false
	f.VisitAll(func(f *flag.Flag) {
		if f.Shorthand == name && f.Value.Type() == "bool" {
			result = true
		}
	})
	return result
}

func stripFlags(args []string, c *Command) []string {
	if len(args) < 1 {
		return args
	}
	c.mergePersistentFlags()

	commands := []string{}

	inQuote := false
	inFlag := false
	for _, y := range args {
		if !inQuote {
			switch {
			case strings.HasPrefix(y, "\""):
				inQuote = true
			case strings.Contains(y, "=\""):
				inQuote = true
			case strings.HasPrefix(y, "--") && !strings.Contains(y, "="):
				// TODO: this isn't quite right, we should really check ahead for 'true' or 'false'
				inFlag = !isBooleanFlag(y[2:], c.Flags())
			case strings.HasPrefix(y, "-") && !strings.Contains(y, "=") && len(y) == 2 && !isBooleanShortFlag(y[1:], c.Flags()):
				inFlag = true
			case inFlag:
				inFlag = false
			case y == "":
				// strip empty commands, as the go tests expect this to be ok....
			case !strings.HasPrefix(y, "-"):
				commands = append(commands, y)
				inFlag = false
			}
		}

		if strings.HasSuffix(y, "\"") && !strings.HasSuffix(y, "\\\"") {
			inQuote = false
		}
	}

	return commands
}

// argsMinusFirstX removes only the first x from args.  Otherwise, commands that look like
// openshift admin policy add-role-to-user admin my-user, lose the admin argument (arg[4]).
func argsMinusFirstX(args []string, x string) []string {
	for i, y := range args {
		if x == y {
			ret := []string{}
			ret = append(ret, args[:i]...)
			ret = append(ret, args[i+1:]...)
			return ret
		}
	}
	return args
}

// find the target command given the args and command tree
// Meant to be run on the highest node. Only searches down.
func (c *Command) Find(args []string) (*Command, []string, error) {
	if c == nil {
		return nil, nil, fmt.Errorf("Called find() on a nil Command")
	}

	var innerfind func(*Command, []string) (*Command, []string)

	innerfind = func(c *Command, innerArgs []string) (*Command, []string) {
		argsWOflags := stripFlags(innerArgs, c)
		if len(argsWOflags) == 0 {
			return c, innerArgs
		}
		nextSubCmd := argsWOflags[0]
		matches := make([]*Command, 0)
		for _, cmd := range c.commands {
			if cmd.Name() == nextSubCmd || cmd.HasAlias(nextSubCmd) { // exact name or alias match
				return innerfind(cmd, argsMinusFirstX(innerArgs, nextSubCmd))
			}
		}

		// only accept a single prefix match - multiple matches would be ambiguous
		if len(matches) == 1 {
			return innerfind(matches[0], argsMinusFirstX(innerArgs, argsWOflags[0]))
		}

		return c, innerArgs
	}

	commandFound, a := innerfind(c, args)
	argsWOflags := stripFlags(a, commandFound)

	// no subcommand, always take args
	if !commandFound.HasSubCommands() {
		return commandFound, a, nil
	}

	// root command with subcommands, do subcommand checking
	if commandFound == c && len(argsWOflags) > 0 {
		return commandFound, a, fmt.Errorf("unknown command %q for %q", argsWOflags[0], commandFound.CommandPath())
	}

	return commandFound, a, nil
}

func (c *Command) Root() *Command {
	var findRoot func(*Command) *Command

	findRoot = func(x *Command) *Command {
		if x.HasParent() {
			return findRoot(x.parent)
		} else {
			return x
		}
	}

	return findRoot(c)
}

// ArgsLenAtDash will return the length of f.Args at the moment when a -- was
// found during arg parsing. This allows your program to know which args were
// before the -- and which came after. (Description from
// https://godoc.org/github.com/spf13/pflag#FlagSet.ArgsLenAtDash).
func (c *Command) ArgsLenAtDash() int {
	return c.Flags().ArgsLenAtDash()
}

func (c *Command) execute(a []string) (err error) {
	if c == nil {
		return fmt.Errorf("Called Execute() on a nil Command")
	}

	// initialize help flag as the last point possible to allow for user
	// overriding
	c.initHelpFlag()

	err = c.ParseFlags(a)
	if err != nil {
		return err
	}
	// If help is called, regardless of other flags, return we want help
	// Also say we need help if the command isn't runnable.
	helpVal, err := c.Flags().GetBool("help")
	if err != nil {
		// should be impossible to get here as we always declare a help
		// flag in initHelpFlag()
		c.Println("\"help\" flag declared as non-bool. Please correct your code")
		return err
	}
	if helpVal || !c.Runnable() {
		return flag.ErrHelp
	}

	argWoFlags := c.Flags().Args()

	for p := c; p != nil; p = p.Parent() {
		if p.PersistentPreRun != nil {
			p.PersistentPreRun(c, argWoFlags)
			break
		}
	}
	if c.PreRun != nil {
		c.PreRun(c, argWoFlags)
	}

	if c.Run != nil {
		c.Run(c, argWoFlags)
	}
	if c.PostRun != nil {
		c.PostRun(c, argWoFlags)
	}
	for p := c; p != nil; p = p.Parent() {
		if p.PersistentPostRun != nil {
			p.PersistentPostRun(c, argWoFlags)
			break
		}
	}

	return nil
}

func (c *Command) errorMsgFromParse() string {
	s := c.flagErrorBuf.String()

	x := strings.Split(s, "\n")

	if len(x) > 0 {
		return x[0]
	} else {
		return ""
	}
}

// Call execute to use the args (os.Args[1:] by default)
// and run through the command tree finding appropriate matches
// for commands and then corresponding flags.
func (c *Command) Execute() (err error) {

	// Regardless of what command execute is called on, run on Root only
	if c.HasParent() {
		return c.Root().Execute()
	}

	// initialize help as the last point possible to allow for user
	// overriding
	c.initHelpCmd()

	var args []string

	if len(c.args) == 0 {
		args = os.Args[1:]
	} else {
		args = c.args
	}

	cmd, flags, err := c.Find(args)
	if err != nil {
		// If found parse to a subcommand and then failed, talk about the subcommand
		if cmd != nil {
			c = cmd
		}
		c.Println("Error:", err.Error())
		c.Printf("Run '%v --help' for usage.\n", c.CommandPath())
		return err
	}

	err = cmd.execute(flags)
	if err != nil {
		if err == flag.ErrHelp {
			cmd.HelpFunc()(cmd, args)
			return nil
		}
		c.Println(cmd.UsageString())
		c.Println("Error:", err.Error())
	}

	return
}

func (c *Command) initHelpFlag() {
	if c.Flags().Lookup("help") == nil {
		c.Flags().BoolP("help", "h", false, "help for "+c.Name())
	}
}

func (c *Command) initHelpCmd() {
	if c.helpCommand == nil {
		if !c.HasSubCommands() {
			return
		}

		c.helpCommand = &Command{
			Use:   "help [command]",
			Short: "Help about any command",
			Long: `Help provides help for any command in the application.
    Simply type ` + c.Name() + ` help [path to command] for full details.`,
			PersistentPreRun:  func(cmd *Command, args []string) {},
			PersistentPostRun: func(cmd *Command, args []string) {},

			Run: func(c *Command, args []string) {
				cmd, _, e := c.Root().Find(args)
				if cmd == nil || e != nil {
					c.Printf("Unknown help topic %#q.", args)
					c.Root().Usage()
				} else {
					helpFunc := cmd.HelpFunc()
					helpFunc(cmd, args)
				}
			},
		}
	}
	c.AddCommand(c.helpCommand)
}

// Used for testing
func (c *Command) ResetCommands() {
	c.commands = nil
	c.helpCommand = nil
}

//Commands returns a slice of child commands.
func (c *Command) Commands() []*Command {
	return c.commands
}

// AddCommand adds one or more commands to this parent command.
func (c *Command) AddCommand(cmds ...*Command) {
	for i, x := range cmds {
		if cmds[i] == c {
			panic("Command can't be a child of itself")
		}
		cmds[i].parent = c
		// update max lengths
		usageLen := len(x.Use)
		if usageLen > c.commandsMaxUseLen {
			c.commandsMaxUseLen = usageLen
		}
		commandPathLen := len(x.CommandPath())
		if commandPathLen > c.commandsMaxCommandPathLen {
			c.commandsMaxCommandPathLen = commandPathLen
		}
		nameLen := len(x.Name())
		if nameLen > c.commandsMaxNameLen {
			c.commandsMaxNameLen = nameLen
		}
		// If glabal normalization function exists, update all children
		if c.globNormFunc != nil {
			x.SetGlobalNormalizationFunc(c.globNormFunc)
		}
		c.commands = append(c.commands, x)
	}
}

// Convenience method to Println to the defined output
func (c *Command) Println(i ...interface{}) {
	fmt.Fprintln(c.Out(), i...)
}

// Convenience method to Printf to the defined output
func (c *Command) Printf(format string, i ...interface{}) {
	fmt.Fprintf(c.Out(), format, i...)
}

// Output the usage for the command
// Used when a user provides invalid input
// Can be defined by user by overriding UsageFunc
func (c *Command) Usage() error {
	c.mergePersistentFlags()
	return c.UsageFunc()(c)
}

// Output the help for the command
// Used when a user calls help [command]
// by the default HelpFunc in the commander
func (c *Command) Help() error {
	c.mergePersistentFlags()
	return tmpl(c.GetOutOr(os.Stdout), c.HelpTemplate(), c)
}

func (c *Command) UsageString() string {
	tmpOutput := c.output
	bb := new(bytes.Buffer)
	c.SetOutput(bb)
	c.Usage()
	c.output = tmpOutput
	return bb.String()
}

// CommandPath returns the full path to this command.
func (c *Command) CommandPath() string {
	str := c.Name()
	x := c
	for x.HasParent() {
		str = x.parent.Name() + " " + str
		x = x.parent
	}
	return str
}

//The full usage for a given command (including parents)
func (c *Command) UseLine() string {
	str := ""
	if c.HasParent() {
		str = c.parent.CommandPath() + " "
	}
	return str + c.Use
}

// For use in determining which flags have been assigned to which commands
// and which persist
func (c *Command) DebugFlags() {
	c.Println("DebugFlags called on", c.Name())
	var debugflags func(*Command)

	debugflags = func(x *Command) {
		if x.HasFlags() || x.HasPersistentFlags() {
			c.Println(x.Name())
		}
		if x.HasFlags() {
			x.flags.VisitAll(func(f *flag.Flag) {
				if x.HasPersistentFlags() {
					if x.persistentFlag(f.Name) == nil {
						c.Println("  -"+f.Shorthand+",", "--"+f.Name, "["+f.DefValue+"]", "", f.Value, "  [L]")
					} else {
						c.Println("  -"+f.Shorthand+",", "--"+f.Name, "["+f.DefValue+"]", "", f.Value, "  [LP]")
					}
				} else {
					c.Println("  -"+f.Shorthand+",", "--"+f.Name, "["+f.DefValue+"]", "", f.Value, "  [L]")
				}
			})
		}
		if x.HasPersistentFlags() {
			x.pflags.VisitAll(func(f *flag.Flag) {
				if x.HasFlags() {
					if x.flags.Lookup(f.Name) == nil {
						c.Println("  -"+f.Shorthand+",", "--"+f.Name, "["+f.DefValue+"]", "", f.Value, "  [P]")
					}
				} else {
					c.Println("  -"+f.Shorthand+",", "--"+f.Name, "["+f.DefValue+"]", "", f.Value, "  [P]")
				}
			})
		}
		c.Println(x.flagErrorBuf)
		if x.HasSubCommands() {
			for _, y := range x.commands {
				debugflags(y)
			}
		}
	}

	debugflags(c)
}

// Name returns the command's name: the first word in the use line.
func (c *Command) Name() string {
	if c.name != "" {
		return c.name
	}
	name := c.Use
	i := strings.Index(name, " ")
	if i >= 0 {
		name = name[:i]
	}
	return name
}

// Determine if a given string is an alias of the command.
func (c *Command) HasAlias(s string) bool {
	for _, a := range c.Aliases {
		if a == s {
			return true
		}
	}
	return false
}

func (c *Command) NameAndAliases() string {
	return strings.Join(append([]string{c.Name()}, c.Aliases...), ", ")
}

func (c *Command) HasExample() bool {
	return len(c.Example) > 0
}

// Determine if the command is itself runnable
func (c *Command) Runnable() bool {
	return c.Run != nil
}

// Determine if the command has children commands
func (c *Command) HasSubCommands() bool {
	return len(c.commands) > 0
}

// IsAvailableCommand determines if a command is available as a non-help command
func (c *Command) IsAvailableCommand() bool {
	if c.HasParent() && c.Parent().helpCommand == c {
		return false
	}

	if c.Runnable() || c.HasAvailableSubCommands() {
		return true
	}

	return false
}

// IsHelpCommand determines if a command is a 'help' command; a help command is
// determined by the fact that it is NOT runnable, and has no
// sub commands that are runnable
func (c *Command) IsHelpCommand() bool {

	// if a command is runnable it is not a 'help' command
	if c.Runnable() {
		return false
	}

	// if any non-help sub commands are found, the command is not a 'help' command
	for _, sub := range c.commands {
		if !sub.IsHelpCommand() {
			return false
		}
	}

	// the command either has no sub commands, or no non-help sub commands
	return true
}

// HasHelpSubCommands determines if a command has any avilable 'help' sub commands
// that need to be shown in the usage/help default template under 'additional help
// topics'
func (c *Command) HasHelpSubCommands() bool {

	// return true on the first found available 'help' sub command
	for _, sub := range c.commands {
		if sub.IsHelpCommand() {
			return true
		}
	}

	// the command either has no sub commands, or no available 'help' sub commands
	return false
}

// HasAvailableSubCommands determines if a command has available sub commands that
// need to be shown in the usage/help default template under 'available commands'
func (c *Command) HasAvailableSubCommands() bool {

	// return true on the first found available (non help)
	// sub command
	for _, sub := range c.commands {
		if sub.IsAvailableCommand() {
			return true
		}
	}

	// the command either has no sub comamnds, or no available (non help)
	// sub commands
	return false
}

// Determine if the command is a child command
func (c *Command) HasParent() bool {
	return c.parent != nil
}

// GlobalNormalizationFunc returns the global normalization function or nil if doesn't exists
func (c *Command) GlobalNormalizationFunc() func(f *flag.FlagSet, name string) flag.NormalizedName {
	return c.globNormFunc
}

// Get the complete FlagSet that applies to this command (local and persistent declared here and by all parents)
func (c *Command) Flags() *flag.FlagSet {
	if c.flags == nil {
		c.flags = flag.NewFlagSet(c.Name(), flag.ContinueOnError)
		if c.flagErrorBuf == nil {
			c.flagErrorBuf = new(bytes.Buffer)
		}
		c.flags.SetOutput(c.flagErrorBuf)
	}
	return c.flags
}

// Get the local FlagSet specifically set in the current command
func (c *Command) LocalFlags() *flag.FlagSet {
	c.mergePersistentFlags()

	local := flag.NewFlagSet(c.Name(), flag.ContinueOnError)
	c.lflags.VisitAll(func(f *flag.Flag) {
		local.AddFlag(f)
	})
	if !c.HasParent() {
		flag.CommandLine.VisitAll(func(f *flag.Flag) {
			if local.Lookup(f.Name) == nil {
				local.AddFlag(f)
			}
		})
	}
	return local
}

// All Flags which were inherited from parents commands
func (c *Command) InheritedFlags() *flag.FlagSet {
	c.mergePersistentFlags()

	inherited := flag.NewFlagSet(c.Name(), flag.ContinueOnError)
	local := c.LocalFlags()

	var rmerge func(x *Command)
	rmerge = func(x *Command) {
		if x.HasPersistentFlags() {
			x.PersistentFlags().VisitAll(func(f *flag.Flag) {
				if inherited.Lookup(f.Name) == nil && local.Lookup(f.Name) == nil {
					inherited.AddFlag(f)
				}
			})
		}
		if x.HasParent() {
			rmerge(x.parent)
		}
	}

	if c.HasParent() {
		rmerge(c.parent)
	}

	return inherited
}

// All Flags which were not inherited from parent commands
func (c *Command) NonInheritedFlags() *flag.FlagSet {
	return c.LocalFlags()
}

// Get the Persistent FlagSet specifically set in the current command
func (c *Command) PersistentFlags() *flag.FlagSet {
	if c.pflags == nil {
		c.pflags = flag.NewFlagSet(c.Name(), flag.ContinueOnError)
		if c.flagErrorBuf == nil {
			c.flagErrorBuf = new(bytes.Buffer)
		}
		c.pflags.SetOutput(c.flagErrorBuf)
	}
	return c.pflags
}

// For use in testing
func (c *Command) ResetFlags() {
	c.flagErrorBuf = new(bytes.Buffer)
	c.flagErrorBuf.Reset()
	c.flags = flag.NewFlagSet(c.Name(), flag.ContinueOnError)
	c.flags.SetOutput(c.flagErrorBuf)
	c.pflags = flag.NewFlagSet(c.Name(), flag.ContinueOnError)
	c.pflags.SetOutput(c.flagErrorBuf)
}

// Does the command contain any flags (local plus persistent from the entire structure)
func (c *Command) HasFlags() bool {
	return c.Flags().HasFlags()
}

// Does the command contain persistent flags
func (c *Command) HasPersistentFlags() bool {
	return c.PersistentFlags().HasFlags()
}

// Does the command has flags specifically declared locally
func (c *Command) HasLocalFlags() bool {
	return c.LocalFlags().HasFlags()
}

func (c *Command) HasInheritedFlags() bool {
	return c.InheritedFlags().HasFlags()
}

// Climbs up the command tree looking for matching flag
func (c *Command) Flag(name string) (flag *flag.Flag) {
	flag = c.Flags().Lookup(name)

	if flag == nil {
		flag = c.persistentFlag(name)
	}

	return
}

// recursively find matching persistent flag
func (c *Command) persistentFlag(name string) (flag *flag.Flag) {
	if c.HasPersistentFlags() {
		flag = c.PersistentFlags().Lookup(name)
	}

	if flag == nil && c.HasParent() {
		flag = c.parent.persistentFlag(name)
	}
	return
}

// Parses persistent flag tree & local flags
func (c *Command) ParseFlags(args []string) (err error) {
	c.mergePersistentFlags()
	err = c.Flags().Parse(args)
	return
}

func (c *Command) Parent() *Command {
	return c.parent
}

func (c *Command) mergePersistentFlags() {
	// Save the set of local flags
	if c.lflags == nil {
		c.lflags = flag.NewFlagSet(c.Name(), flag.ContinueOnError)
		if c.flagErrorBuf == nil {
			c.flagErrorBuf = new(bytes.Buffer)
		}
		c.lflags.SetOutput(c.flagErrorBuf)
		addtolocal := func(f *flag.Flag) {
			c.lflags.AddFlag(f)
		}
		c.Flags().VisitAll(addtolocal)
		c.PersistentFlags().VisitAll(addtolocal)
	}

	var rmerge func(x *Command)
	rmerge = func(x *Command) {
		if !x.HasParent() {
			flag.CommandLine.VisitAll(func(f *flag.Flag) {
				if x.PersistentFlags().Lookup(f.Name) == nil {
					x.PersistentFlags().AddFlag(f)
				}
			})
		}
		if x.HasPersistentFlags() {
			x.PersistentFlags().VisitAll(func(f *flag.Flag) {
				if c.Flags().Lookup(f.Name) == nil {
					c.Flags().AddFlag(f)
				}
			})
		}
		if x.HasParent() {
			rmerge(x.parent)
		}
	}

	rmerge(c)
}
