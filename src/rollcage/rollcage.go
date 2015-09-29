// rollcage
package main

import "rollcage/commands"

var Version = "no-version"

func main() {
	commands.Version = Version
	commands.RootCmd.Execute()
}
