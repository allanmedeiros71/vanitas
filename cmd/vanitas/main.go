// Command vanitas is a terminal database client.
package main

import (
	"os"

	"github.com/allanmedeiros71/vanitas/internal/cli"
)

func main() {
	os.Exit(cli.Execute())
}
