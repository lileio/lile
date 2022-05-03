package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	_ "github.com/lileio/lile/v2/statik"
	"github.com/mattn/go-colorable"
	"github.com/spf13/cobra"
)

var (
	dir  string
	name string

	out    = colorable.NewColorableStdout()
	newCmd = &cobra.Command{
		Use:   "new [name]",
		Short: "Create a new service",
		Run:   new,
	}
)

func init() {
	RootCmd.AddCommand(newCmd)

	newCmd.Flags().StringVar(
		&dir,
		"dir",
		"",
		"the directory to create the service",
	)

	newCmd.Flags().StringVar(
		&name,
		"name",
		"",
		"the module name i.e (github.com/username/project)",
	)

	newCmd.MarkFlagRequired("name")
}

func new(cmd *cobra.Command, args []string) {
	if dir == "" {
		dir = lastFromSplit(name, string(os.PathSeparator))
	}

	fmt.Printf("Creating project in %s\n", dir)

	if !askIsOK() {
		fmt.Println("Exiting..")
		return
	}

	p := newProject(dir, name)

	err := p.write()
	if err != nil {
		er(err)
	}

	p.Folder.print()
}

func askIsOK() bool {
	if os.Getenv("CI") != "" {
		return true
	}

	fmt.Fprintf(out, "Is this OK? %ses/%so\n",
		color.YellowString("[y]"),
		color.CyanString("[n]"),
	)
	scan := bufio.NewScanner(os.Stdin)
	scan.Scan()
	return strings.Contains(strings.ToLower(scan.Text()), "y")
}

func er(err error) {
	if err != nil {
		fmt.Fprintf(out, "%s: %s \n",
			color.RedString("[ERROR]"),
			err.Error(),
		)
		panic(err)
	}
}
