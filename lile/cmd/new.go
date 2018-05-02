package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new [name]",
	Short: "Create a new microservice",
	Run:   new,
}

var (
	gopath       string
	templatePath string
)

var out = colorable.NewColorableStdout()

func init() {
	gopath = os.Getenv("GOPATH")
	if gopath == "" {
		b, err := exec.Command("go", "env", "GOPATH").CombinedOutput()
		if err != nil {
			panic(string(b))
		}
		gopath = strings.TrimSpace(string(b))
	}

	if paths := filepath.SplitList(gopath); len(paths) > 0 {
		gopath = paths[0]
	}

	templatePath = filepath.Clean(filepath.Join(gopath, "/src/github.com/lileio/lile/template"))
	RootCmd.AddCommand(newCmd)
}

func new(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		fmt.Printf("You must supply a path for the service, e.g lile new lile/users\n")
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
