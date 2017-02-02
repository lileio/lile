package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/serenize/snaker"
)

type Project struct {
	Name       string
	ProjectDir string
	RelDir     string
	Folder     folder
}

func NewProject(path string) Project {
	name := lastFromSplit(path, string(os.PathSeparator))
	relDir := projectBase(path)

	f := folder{Name: name, AbsPath: path}
	s := f.addFolder("server")
	s.addFile("server.go", "server.tmpl")
	s.addFile("server_test.go", "server_test.tmpl")
	f.addFolder(name).addFile(name+".proto", "proto.tmpl")
	f.addFile("Makefile", "Makefile.tmpl")
	f.addFile("readme.md", "readme.tmpl")
	f.addFile("main.go", "main.tmpl")
	f.addFile("wercker.yml", "wercker.tmpl")

	return Project{
		Name:       name,
		RelDir:     relDir,
		ProjectDir: path,
		Folder:     f,
	}
}

func (p Project) Write(templatePath string) error {
	err := os.Mkdir(p.ProjectDir, os.ModePerm)
	if err != nil {
		return err
	}

	return p.Folder.render(templatePath, p)
}

func (p Project) CamelCaseName() string {
	return snaker.SnakeToCamel(p.Name)
}

func (p Project) SnakeCaseName() string {
	return snaker.CamelToSnake(p.Name)
}

// Copied and re-worked from
// https://github.com/spf13/cobra/blob/master/cobra/cmd/helpers.go
func projectPath(inputPath string) string {
	// if no path is provided... assume CWD.
	if inputPath == "" {
		x, err := os.Getwd()
		if err != nil {
			er(err)
		}

		return x
	}

	var projectPath string
	var projectBase string
	srcPath := srcPath()

	// if provided, inspect for logical locations

	if strings.ContainsRune(inputPath, os.PathSeparator) {
		if filepath.IsAbs(inputPath) || filepath.HasPrefix(inputPath, string(os.PathSeparator)) {
			// if Absolute, use it
			projectPath = filepath.Clean(inputPath)
			return projectPath
		}
		// If not absolute but contains slashes,
		// assuming it means create it from $GOPATH
		count := strings.Count(inputPath, string(os.PathSeparator))

		switch count {
		// If only one directory deep, assume "github.com"
		case 1:
			projectPath = filepath.Join(srcPath, "github.com", inputPath)
			return projectPath
		case 2:
			projectPath = filepath.Join(srcPath, inputPath)
			return projectPath
		default:
			er(errors.New("Unknown directory"))
		}
	}

	// hardest case.. just a word.
	if projectBase == "" {
		x, err := os.Getwd()
		if err == nil {
			projectPath = filepath.Join(x, inputPath)
			return projectPath
		}
		er(err)
	}

	projectPath = filepath.Join(srcPath, projectBase, inputPath)
	return projectPath
}

func projectBase(absPath string) string {
	return lastFromSplit(absPath, srcPath())
}

func lastFromSplit(input, split string) string {
	rel := strings.Split(input, split)
	return rel[len(rel)-1]
}

func srcPath() string {
	return filepath.Join(os.Getenv("GOPATH"), "src") + string(os.PathSeparator)
}
