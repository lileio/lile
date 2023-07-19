package cmd

import (
	"bytes"
	"fmt"
	"go/format"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/rakyll/statik/fs"
	"github.com/xlab/treeprint"

	_ "github.com/lileio/lile/v2/statik"
)

type file struct {
	Name     string
	AbsPath  string
	Template string
}

type folder struct {
	AbsPath string

	// Unexported so you can't set them without methods
	files   []file
	folders []*folder
}

func (f *folder) addFolder(name string) *folder {
	newF := &folder{
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

func (f *folder) render(p project) error {
	hfs, err := fs.New()
	if err != nil {
		log.Fatal(err)
	}

	for _, v := range f.files {
		b, err := fs.ReadFile(hfs, "/"+v.Template)
		if err != nil {
			return err
		}

		t, err := template.New(v.Template).Parse(string(b))
		if err != nil {
			return err
		}

		file, err := os.Create(v.AbsPath)
		if err != nil {
			return err
		}

		defer file.Close()

		if strings.Contains(v.AbsPath, ".go") {
			var out bytes.Buffer
			err = t.Execute(&out, p)
			if err != nil {
				log.Printf("Could not process template %s\n", v)
				return err
			}

			b, err := format.Source(out.Bytes())
			if err != nil {
				fmt.Print(out.String())
				log.Printf("\nCould not format Go file %s\n", v)
				return err
			}

			_, err = file.Write(b)
			if err != nil {
				return err
			}
		} else {
			err = t.Execute(file, p)
			if err != nil {
				return err
			}
		}
	}

	for _, v := range f.folders {
		err := os.Mkdir(v.AbsPath, os.ModePerm)
		if err != nil {
			return err
		}

		err = v.render(p)
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
		tree = tree.AddBranch(lastFromSplit(f.AbsPath, string(os.PathSeparator)))
	}

	for _, v := range f.folders {
		v.tree(false, tree)
	}

	for _, v := range f.files {
		tree.AddNode(v.Name)
	}

	return tree
}
