# Lile

#  [![Go Report Card](https://goreportcard.com/badge/github.com/lileio/lile)](https://goreportcard.com/report/github.com/lileio/lile) [![wercker status](https://app.wercker.com/status/655c3bdcdeb2334335fda4959f3ad5cb/s/master "wercker status")](https://app.wercker.com/project/byKey/655c3bdcdeb2334335fda4959f3ad5cb) [![GoDoc](https://godoc.org/github.com/lileio/lile?status.svg)](https://godoc.org/github.com/lileio/lile)

Lile is a generator and library to help you build [gRPC](grpc.io) based services easily. It in cludes default metrics and tracing out of the box and be configured if needed.

A service build in Lile can easily remove Lile and just use gRPC's server at a later time if needed, the Lile library only wraps gRPCs server to get you up and running quickly.

![](http://g.recordit.co/0kUVNorbsZ.gif)

## Getting Started

### Creating a new service

To work with Lile you'll need a working Go environment and protobuf/grpc installed, see instructions [here](http://www.grpc.io/docs/quickstart/go.html).

Install Lile by running:

```Bash
go get github.com/lileio/lile/lile
```

Then you can generate your new gRPC service by calling

```Bash
lile new github_username/user_service
```

Lile is clever enough to find the right GOPATH and will generate the app into the path `$GOPATH/github.com/github_username/user_service`

This will generate a new 'project' for your service will the following structure:

``` 
├── server
│   ├── server.go
│   └── server_test.go
├── user_service
│   └── user_service.proto
├── Makefile
├── readme.md
├── main.go
└── wercker.yml
```

You can then generate the proto and run the test suite!

```
make proto
make test
```

Hopefully you have a working test suite!