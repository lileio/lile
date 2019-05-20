package cmd

import (
	"os"
	"strings"

	"github.com/iancoleman/strcase"
)

type project struct {
	ModuleName string
	Name       string
	ProjectDir string
	Folder     folder
}

func newProject(path, moduleName string) project {
	f := folder{
		AbsPath: path,
	}

	name := lastFromSplit(name, string(os.PathSeparator))

	s := f.addFolder("server")
	s.addFile("server.go", "server.tmpl")
	s.addFile("server_test.go", "server_test.tmpl")

	subs := f.addFolder("subscribers")
	subs.addFile("subscribers.go", "subscribers.tmpl")

	cmd := f.addFolder(name)
	cmd.addFile("main.go", "cmd_main.tmpl")

	cmds := cmd.addFolder("cmd")
	cmds.addFile("root.go", "cmd_root.tmpl")
	cmds.addFile("up.go", "cmd_up.tmpl")

	f.addFile(name+".proto", "proto.tmpl")
	f.addFile("client.go", "client.tmpl")
	f.addFile("Makefile", "Makefile.tmpl")
	f.addFile("Dockerfile", "Dockerfile.tmpl")
	f.addFile("go.mod", "go-mod.tmpl")
	f.addFile(".gitignore", "gitignore.tmpl")

	return project{
		ModuleName: moduleName,
		Name:       name,
		ProjectDir: path,
		Folder:     f,
	}
}

func (p project) write() error {
	err := os.MkdirAll(p.ProjectDir, os.ModePerm)
	if err != nil {
		return err
	}

	return p.Folder.render(p)
}

// CamelCaseName returns a CamelCased name of the service
func (p project) CamelCaseName() string {
	return strcase.ToCamel(p.Name)
}

// DNSName returns a snake-cased-type name that be used as a URL or packageName
func (p project) DNSName() string {
	return strings.Replace(strcase.ToSnake(p.Name), "_", "-", -1)
}

func lastFromSplit(input, split string) string {
	rel := strings.Split(input, split)
	return rel[len(rel)-1]
}
