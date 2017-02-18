package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new [name]",
	Short: "Create a new microservice",
	Run:   new,
}

var templatePath = os.Getenv("GOPATH") + "/src/github.com/lileio/lile/template"

func init() {
	RootCmd.AddCommand(newCmd)
}

func new(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		fmt.Printf("You must supply a path for the service, e.g lile new lile/user_service\n")
		return
	}

	name := args[0]
	path := projectPath(name)
	fmt.Printf("Creating project in %s\n", path)

	if !askIsOK() {
		fmt.Println("Exiting..")
		return
	}

	p := newProject(path, name)

	err := p.write(templatePath)
	if err != nil {
		er(err)
	}

	p.Folder.print()
}

func askIsOK() bool {
	if os.Getenv("CI") != "" {
		return true
	}

	fmt.Printf("Is this OK? %ses/%so\n",
		color.YellowString("[y]"),
		color.CyanString("[n]"),
	)
	scan := bufio.NewScanner(os.Stdin)
	scan.Scan()
	return strings.Contains(strings.ToLower(scan.Text()), "y")
}

func er(err error) {
	if err != nil {
		fmt.Printf("%s: %s \n",
			color.RedString("[ERROR]"),
			err.Error(),
		)
		os.Exit(-1)
	}
}
