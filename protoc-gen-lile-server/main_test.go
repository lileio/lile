package main

import (
	"bytes"
	"log"
	"testing"

	"github.com/golang/protobuf/proto"
	protodescriptor "github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/stretchr/testify/assert"
)

func TestGenerator(t *testing.T) {
	req := &plugin.CodeGeneratorRequest{
		FileToGenerate: []string{"example.proto"},
		ProtoFile: []*protodescriptor.FileDescriptorProto{
			{
				Name:    proto.String("empty.proto"),
				Package: proto.String("google.protobuf"),
				Options: &protodescriptor.FileOptions{
					GoPackage: proto.String("github.com/golang/protobuf/ptypes/empty"),
				},
			},
			stubFile(),
		},
	}

	data, err := proto.Marshal(req)
	if err != nil {
		log.Fatal("marshaling error: ", err)
	}

	r := bytes.NewReader(data)
	w := &bytes.Buffer{}
	gen(r, w)

	var res plugin.CodeGeneratorResponse
	err = proto.Unmarshal(w.Bytes(), &res)
	if err != nil {
		log.Fatal("unmarshaling error: ", err)
	}

	assert.Nil(t, res.Error)
	assert.Equal(t, len(res.File), 10)
}

func stubFile() *protodescriptor.FileDescriptorProto {
	msgdesc := &protodescriptor.DescriptorProto{
		Name: proto.String("ExampleMessage"),
	}

	unary_meth := &protodescriptor.MethodDescriptorProto{
		Name:       proto.String("Example"),
		InputType:  proto.String("example.ExampleMessage"),
		OutputType: proto.String("example.ExampleMessage"),
	}

	custom_type_meth := &protodescriptor.MethodDescriptorProto{
		Name:       proto.String("ExampleWithoutBindings"),
		InputType:  proto.String("google.protobuf.Empty"),
		OutputType: proto.String("google.protobuf.Empty"),
	}

	unary_stream_meth := &protodescriptor.MethodDescriptorProto{
		Name:            proto.String("Example"),
		InputType:       proto.String("example.ExampleMessage"),
		OutputType:      proto.String("example.ExampleMessage"),
		ServerStreaming: proto.Bool(true),
	}

	stream_unary_meth := &protodescriptor.MethodDescriptorProto{
		Name:            proto.String("Example"),
		InputType:       proto.String("example.ExampleMessage"),
		OutputType:      proto.String("example.ExampleMessage"),
		ClientStreaming: proto.Bool(true),
	}

	stream_stream_meth := &protodescriptor.MethodDescriptorProto{
		Name:            proto.String("Example"),
		InputType:       proto.String("example.ExampleMessage"),
		OutputType:      proto.String("example.ExampleMessage"),
		ClientStreaming: proto.Bool(true),
		ServerStreaming: proto.Bool(true),
	}

	svc := &protodescriptor.ServiceDescriptorProto{
		Name: proto.String("ExampleService"),
		Method: []*protodescriptor.MethodDescriptorProto{
			unary_meth,
			custom_type_meth,
			unary_stream_meth,
			stream_unary_meth,
			stream_stream_meth,
		},
	}

	gopkg := "github.com/example/example"
	return &protodescriptor.FileDescriptorProto{
		Name:        proto.String("example.proto"),
		Options:     &protodescriptor.FileOptions{GoPackage: &gopkg},
		Package:     proto.String("example"),
		Dependency:  []string{"google/protobuf/empty.proto"},
		MessageType: []*protodescriptor.DescriptorProto{msgdesc},
		Service:     []*protodescriptor.ServiceDescriptorProto{svc},
	}
}
