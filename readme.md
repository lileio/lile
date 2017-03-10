# ![logo](https://cdn.rawgit.com/lileio/lile/aa45e6ae200692b4668bc6e370e1757e7753a514/logo.svg)

[![Go Report Card](https://goreportcard.com/badge/github.com/lileio/lile)](https://goreportcard.com/report/github.com/lileio/lile) [![Build Status](https://travis-ci.org/lileio/lile.svg?branch=master)](https://travis-ci.org/lileio/lile) [![GoDoc](https://godoc.org/github.com/lileio/lile?status.svg)](https://godoc.org/github.com/lileio/lile)

Lile is a app generator and utility library to help you build [gRPC](http://grpc.io) based services easily in Go with other languages hopefully coming soon. 

The Lile generator creates a ready to go service with features like a working test suite, Dockerfile, Makefile and a cmd line based app amongst others.

The utility library extends the basic gRPC server with nice defaults like metrics, tracing and logging.

A service built with Lile can easily remove Lile as a dependency as Lile is NOT a framework, just a set of helpers.

# ![logo](https://dl.dropboxusercontent.com/u/7788162/lile.png)

## Lileio

Since Lile services are gRPC services they can be consumed in many languages and clients can be generated automatically.

Mobile apps can communicate with gRPC and browser support is being actively [worked on.](https://github.com/grpc/grpc/issues/8682) 

RESTful JSON can be automatically be generated by using the [gRPC gateway](https://github.com/grpc-ecosystem/grpc-gateway), you can even provide Swagger documentation. See [here](https://coreos.com/blog/grpc-protobufs-swagger.html) for how CoreOS combine a REST, Swagger and gRPC server.

## Getting Started

### Creating a new service

To work with Lile you'll need a working Go environment and protobuf/grpc installed, see instructions [here](http://www.grpc.io/docs/quickstart/go.html).

Install Lile by running:

```Bash
go get -u github.com/lileio/lile/lile
```

Then you can generate your new gRPC service by calling

```Bash
lile new github_username/user_service
```

Lile is clever enough to find the right GOPATH and will generate the app into the path `$GOPATH/github.com/github_username/user_service`

This will generate a new 'project' for your service will the following structure:

``` 
.
├── server
│   ├── server.go
│   └── server_test.go
├── user_service
│   ├── cmd
│       ├── root.go
│       └── server.go
│   └── main.go
├── user_service.proto
├── Makefile
├── Dockerfile
├── readme.md
├── .travis.yml
└── .gitignore
```

You can then generate the proto and run the test suite!

```
go get -u github.com/golang/protobuf/{proto,protoc-gen-go}
make proto
make get
make test
```

Hopefully you have a working test suite!

To run the service run the cmd line based server

```
go run ./user_service/main.go
```

# Guide

- [Feedback](#feedback-welcome)
- [Makefile](#makefile)
  - [proto](#proto)
  - [run](#run)
  - [test](#test)
  - [ci](#ci)
  - [benchmark](#benchmark)
  - [docker](#docker)
- [Travis](#travis)
- [Monitoring](#monitoring)
  - [Prometheus](#prometheus)
- [Tracing](#tracing)
  - [Zipkin via HTTP](#zipkin-via-http)
  - [Zipkin via Kafka](#zipkin-via-kafka)
  - [opentracing](#opentracing)

## Feedback welcome!

Lile and it's templates are up for discussion! I'm keen to take PR's and start discussions around the contents and code generated. It'd be great to create an open source collection of awesome services we could all borrow and use.

## Makefile

Lile generates a [Makefile](http://mrbook.org/blog/tutorials/make/) as part of your service. It's filled with some default commands but you can of course change them or add your own.

#### proto

`make proto` will run the `protoc` compiler with the grpc go plugin to create a go packge that you gRPC uses as the server. It also creates a client that other Go programs can import and use! See [here](http://www.grpc.io/docs/quickstart/go.html) for more info.

#### test

`make test` runs `go test -v ./…` which simply means "run all tests including sub packages in verbose mode". I have `make test` aliased to `t` on my machine.

#### ci

`make ci` runs `make get` and make `test`

#### benchmark

`make benchmark` is the same behaviour as `make test` but runs the Go benchmark suite, the default time is 10s.

#### docker

`make docker` cross compiles into a linux arch64 app and then copies that over the binary and code (in case you have templates etc) to an arch linux container by default.

## Travis

Lile generates a Travis CI yaml file which runs the just test suite by default. 

Support is included for pushing to a docker repository on the master branch after a successful build, but needs uncommenting.

## Monitoring

Lile includes built in app monitoring, which by default is prometheus.

### Prometheus

By default Lile will add the Prometheus gRPC interceptor and collect gRPC metrics.

You can access these at `/metrics` on port `8080` or change them like below:

``` go
lile.NewServer(
  lile.PrometheusPort(":9999"),
  lile.PrometheusAddr("/prom"),
)
```

You can disable prometheus interception by setting `PrometheusEnabled` to `false`:

``` go
lile.NewServer(
  lile.PrometheusEnabled(false),
)
```

## Tracing

Lile support opentracing tracers and ships with [Zipkin](http://zipkin.io/) compatiable tracing by default. Though you will need to set an ENV var to enable the feature.

Lile sets the gRPC tracer when the tracing option is set, below is [Google Cloud Tracing](https://cloud.google.com/trace/) collected via the [stackdriver zipkin](https://github.com/GoogleCloudPlatform/stackdriver-zipkin) Docker container

# ![trace](https://dl.dropboxusercontent.com/u/7788162/trace.png)

#### Zipkin via HTTP

To have Lile send traces to Zipkin via the HTTP endpoint set the `ZIPKIN_HTTP_ENDPOINT` env variable

`ZIPKIN_HTTP_ENDPOINT=http://localhost:9411/api/v1/spans`

#### Zipkin via Kafka

To have Lile send traces to Zipkin via Kakfa (to then be collected with a Zipkin collector) set the `ZIPKIN_KAFKA_ENDPOINTS` env variable. 

It can be a single Kafka host or multiple hosts seperated by comma

`ZIPKIN_KAFKA_ENDPOINTS=10.0.0.1:9092,10.0.0.2:9092`

#### Opentracing

Lile supports opentracing compatible tracers via [otgrpc](https://github.com/grpc-ecosystem/grpc-opentracing/tree/master/go/otgrpc) and [opentracing-go](https://github.com/opentracing/opentracing-go)

You can set a tracer using the `Tracer` option..

``` go
lile.NewServer(
  lile.Tracer(sometracer),
)
```

You can disable tracing completely using the `TracingEnabled` option

``` go
lile.NewServer(
  lile.TracingEnabled(false),
)
```