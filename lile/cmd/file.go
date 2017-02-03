package cmd

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"github.com/xlab/treeprint"
)

type file struct {
	Name     string
	AbsPath  string
	Template string
}

type folder struct {
	Name    string
	AbsPath string

	// Unexported so you can't set them without methods
	files   []file
	folders []*folder
}

func (f *folder) addFolder(name string) *folder {
	newF := &folder{
		Name:    name,
		AbsPath: filepath.Join(f.AbsPath, name),
	}
	f.folders = append(f.folders, newF)
	return newF
}

func (f *folder) addFile(name, tmpl string) {
	f.files = append(f.files, file{
		Name:     name,
		Template: tmpl,
		AbsPath:  filepath.Join(f.AbsPath, name),
	})
}

func (f *folder) render(templatePath string, p project) error {
	for _, v := range f.files {
		t, err := template.ParseFiles(filepath.Join(templatePath, v.Template))
		if err != nil {
			return err
		}

		file, err := os.Create(v.AbsPath)
		if err != nil {
			return err
		}

		err = t.Execute(file, p)
		if err != nil {
			return err
		}
	}

	for _, v := range f.folders {
		err := os.Mkdir(v.AbsPath, os.ModePerm)
		if err != nil {
			return err
		}

		err = v.render(templatePath, p)
		if err != nil {
			return err
		}
	}

	return nil
}

func (f folder) print() {
	t := f.tree(true, treeprint.New())
	fmt.Println(t.String())
}

func (f folder) tree(root bool, tree treeprint.Tree) treeprint.Tree {
	if !root {
		tree = tree.AddBranch(f.Name)
	}

	for _, v := range f.folders {
		v.tree(false, tree)
	}

	for _, v := range f.files {
		tree.AddNode(v.Name)
	}

	return tree
}
