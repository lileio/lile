package main

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"text/template"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/serenize/snaker"
	"github.com/xtgo/set"

	"github.com/fatih/color"
	"github.com/golang/glog"
	"github.com/golang/protobuf/proto"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
)

var (
	gopath       string
	templatePath string

	input  io.Reader
	output io.Writer
)

type grpcMethod struct {
	ServiceName     string
	ImportName      string
	Name            string
	InType          string
	OutType         string
	ClientStreaming bool
	ServerStreaming bool

	GoPackage    string
	InputImport  string
	OutputImport string
}

// Used to import other packages correctly, for example..
// Package: "google.protobuf"
// GoPackage: "github.com/golang/protobuf/ptypes/empty"
// we replace all occurences of Package with the end of GoPackage and import it
type goimport struct {
	Package   string
	GoPackage string
	GoType    string
}

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
	templatePath = filepath.Clean(filepath.Join(gopath, "/src/github.com/lileio/lile/protoc-gen-lile-server/templates"))
}

func main() {
	gen(os.Stdin, os.Stdout)
}

func gen(i io.Reader, o io.Writer) {
	input = i
	output = o

	// Force color output
	if runtime.GOOS != "windows" {
		color.NoColor = false
	}

	// Parse the incoming protobuf request
	req, err := parseReq(input)
	if err != nil {
		emitError(err)
		log.Fatal(err)
	}

	path := "./server"
	if req.Parameter != nil {
		path = *req.Parameter
	}

	files := []*plugin.CodeGeneratorResponse_File{}
	imports := []goimport{}

	for _, file := range req.ProtoFile {
		pkgParts := strings.Split(file.GetOptions().GetGoPackage(), "/")
		imports = append(imports, goimport{
			Package:   file.GetPackage(),
			GoPackage: file.GetOptions().GetGoPackage(),
			GoType:    pkgParts[len(pkgParts)-1],
		})

		// Only generate methods for this file/project
		if *file.Name != req.FileToGenerate[0] {
			if file.Options == nil || file.Options.GoPackage == nil {
				log.Fatalf("No go_package option defined for import %s", *file.Name)
			}

			continue
		}

		if file.Options == nil || file.Options.GoPackage == nil {
			log.Fatalf("No go_package option defined in %s", *file.Name)
		}

		pkgSplit := strings.Split(file.Options.GetGoPackage(), "/")

		for _, service := range file.Service {
			for _, method := range service.Method {
				gm := grpcMethod{
					ServiceName:     service.GetName(),
					GoPackage:       file.Options.GetGoPackage(),
					ImportName:      pkgSplit[len(pkgSplit)-1],
					Name:            method.GetName(),
					InType:          toGoType(imports, method.GetInputType()),
					OutType:         toGoType(imports, method.GetOutputType()),
					ClientStreaming: method.GetClientStreaming(),
					ServerStreaming: method.GetServerStreaming(),

					InputImport:  inputImport(imports, method),
					OutputImport: outputImport(imports, method),
				}

				f, err := generateMethod(path, gm)
				if err != nil {
					emitError(err)
					log.Fatal(err)
				}

				if len(f) > 0 {
					for _, v := range f {
						files = append(files, v)
					}
				}
			}
		}
	}

	emitFiles(files)
}

func DedupImports(imports ...string) string {
	data := sort.StringSlice(imports)
	sort.Sort(data)
	n := set.Uniq(data)
	imports = data[:n]

	return fmt.Sprintf("\t\"%s\"", strings.Join(imports, "\"\n\t\""))
}

func toGoType(imports []goimport, t string) string {
	t = strings.Trim(t, ".")
	for _, i := range imports {
		if strings.Contains(t, i.Package) {
			s := strings.Replace(t, i.Package, i.GoType, 1)
			s = strings.Replace(s, "-", "_", -1)
			return s
		}
	}

	return t
}

func inputImport(imports []goimport, method *descriptor.MethodDescriptorProto) string {
	for _, i := range imports {
		if strings.Contains(method.GetInputType(), i.Package) {
			return i.GoPackage
		}
	}

	return ""
}

func outputImport(imports []goimport, method *descriptor.MethodDescriptorProto) string {
	for _, i := range imports {
		if strings.Contains(method.GetOutputType(), i.Package) {
			return i.GoPackage
		}
	}

	return ""
}

func generateMethod(basePath string, m grpcMethod) ([]*plugin.CodeGeneratorResponse_File, error) {
	path := filepath.Join(basePath, snaker.CamelToSnake(m.Name)+".go")
	test_path := filepath.Join(basePath, snaker.CamelToSnake(m.Name)+"_test.go")

	template := fmt.Sprintf("%s_%s",
		streamFromBool(m.ClientStreaming),
		streamFromBool(m.ServerStreaming),
	)

	// If the file exists then just skip the creation
	_, err := os.Stat(path)
	if err == nil {
		log.Printf("%s %s", color.YellowString("[Skipping]"), path)
		return nil, nil
	}

	files := []*plugin.CodeGeneratorResponse_File{}

	log.Printf("%s %s", color.GreenString("[Creating]"), path)
	f, err := render(path, template+".tmpl", m)
	if err != nil {
		return nil, err
	}

	files = append(files, f)

	log.Printf("%s %s", color.CyanString("[Creating test]"), test_path)
	f, err = render(test_path, template+"_test.tmpl", m)

	files = append(files, f)
	return files, err
}

func streamFromBool(streaming bool) string {
	if streaming {
		return "stream"
	}

	return "unary"
}

func render(path, tmpl string, m grpcMethod) (*plugin.CodeGeneratorResponse_File, error) {
	t := template.New(tmpl)

	funcMap := template.FuncMap{
		"dedupImports": DedupImports,
	}

	t = t.Funcs(funcMap)
	t, err := t.ParseFiles(filepath.Join(templatePath, tmpl))
	if err != nil {
		return nil, err
	}

	var out bytes.Buffer
	err = t.Execute(&out, m)
	if err != nil {
		log.Printf("%s couldn't create template %s, %s", color.RedString("[ERROR]"), tmpl, err)
		return nil, err
	}

	b, err := format.Source(out.Bytes())
	if err != nil {
		log.Printf(string(out.Bytes()))
		log.Printf("\n%s couldn't format Go file %s, %s", color.RedString("[ERROR]"), tmpl, err)
		return nil, err
	}

	str := string(b)

	return &plugin.CodeGeneratorResponse_File{
		Name:    &path,
		Content: &str,
	}, nil
}

func parseReq(r io.Reader) (*plugin.CodeGeneratorRequest, error) {
	input, err := ioutil.ReadAll(r)
	if err != nil {
		log.Fatalf("Failed to read code generator request: %v", err)
		return nil, err
	}

	req := new(plugin.CodeGeneratorRequest)
	if err = proto.Unmarshal(input, req); err != nil {
		log.Fatalf("Failed to unmarshal code generator request: %v", err)
		return nil, err
	}

	return req, nil
}

func emitFiles(files []*plugin.CodeGeneratorResponse_File) {
	emitResp(&plugin.CodeGeneratorResponse{File: files})
}

func emitError(err error) {
	emitResp(&plugin.CodeGeneratorResponse{Error: proto.String(err.Error())})
}

func emitResp(resp *plugin.CodeGeneratorResponse) {
	buf, err := proto.Marshal(resp)
	if err != nil {
		glog.Fatal(err)
	}
	if _, err := output.Write(buf); err != nil {
		glog.Fatal(err)
	}
}
