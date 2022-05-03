package main

import (
	"fmt"
	"os"

	"github.com/lileio/lile/v2/lile/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
