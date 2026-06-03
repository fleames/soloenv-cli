// Command soloenv gives solo developers a shareable public staging URL for
// their existing Docker app from a single command, then tears it down on demand.
package main

import "github.com/fleames/soloenv-cli/cmd"

func main() {
	cmd.Execute()
}
