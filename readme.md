
# ![logo](https://cdn.rawgit.com/lileio/lile/aa45e6ae200692b4668bc6e370e1757e7753a514/logo.svg)

[![Go Report Card](https://goreportcard.com/badge/github.com/lileio/lile)](https://goreportcard.com/report/github.com/lileio/lile) [![Build Status](https://travis-ci.org/lileio/lile.svg?branch=master)](https://travis-ci.org/lileio/lile) [![GoDoc](https://godoc.org/github.com/lileio/lile?status.svg)](https://godoc.org/github.com/lileio/lile)

Lile is a generator and library to help you build [gRPC](http://grpc.io) based services easily in Go (other languages coming soon). It includes default metrics and tracing out of the box and be configured if needed.

A service built with Lile can easily remove Lile as a dependency later as a Lile service just speaks gRPC.

![](http://g.recordit.co/0kUVNorbsZ.gif)

## Lileio

The lile project aims to write and maintain open source services as building blocks for others, so you don't have to write an "accounts" or "auth" service every time!

See the Lile [Github account](https://github.com/lileio/) for a list of services and feel free to write and contribute services!

It's important to note that gRPC services don't have to be "micro" and can be consumed in many languages. Mobile apps can communicate with gRPC and browser support is being actively [worked on.](https://github.com/grpc/grpc/issues/8682) Note that a lot of services are intended to be consumed by __your__ code, i.e behind a firewall, with an "API gateway" being the main communication with the outside world. (in the same way that you wouldn't expose your database)

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
├── Dockerfile
├── readme.md
├── main.go
└── .travis.yml
```

You can then generate the proto and run the test suite!

```
go get -u github.com/golang/protobuf/{proto,protoc-gen-go}
make proto
make test
```

Hopefully you have a working test suite!

# Guide

- [Feedback](#feedback-welcome)
- [Makefile](#makefile)
  - [proto](#proto)
  - [run](#run)
  - [test](#test)
  - [benchmark](#benchmark)
  - [docker](#docker)
- [Travis](#travis)
- [Monitoring](#monitoring)
  - [Prometheus](#prometheus)
- [Tracing](#tracing)
  - [Zipkin via HTTP](#zipkin-via-http)
  - [Zipkin via Scribe](#zipkin-via-scribe)
  - [Zipkin via Kafka](#zipkin-via-kafka)
  - [opentracing](#opentracing)

## Feedback welcome!

Lile and it's templates are up for discussion! I'm keen to take PR's and start discussions around the contents and code produced and would love to work on a community concencous of best practices. For example, if you feel Make isn't the way forward and think we should use something else, let's discuss. 

## Makefile

Lile generate a [Makefile](http://mrbook.org/blog/tutorials/make/) as part of the project. It's filled with some default commands but you can of course change them or add your own.

#### proto

`make proto` will run the `protoc` compiler with the grpc go plugin to create a go packge that you gRPC uses as the server. It also creates a client that other Go programs can import and use! See [here](http://www.grpc.io/docs/quickstart/go.html) for more info.

#### run

`make run` is a simple shortcut to `go run main.go`, it's only really there because it's a little shorter to type. Especially since I for example have `make` aliased to `m` on my machine. Therefore `m run` runs my Go programs.

#### test

`make test` runs `go test -v ./…` which simply means "run all tests including sub packages in verbose mode". I have `make test` aliased to `t` on my machine.

#### benchmark

`make benchmark` is the same behaviour as `make test` but runs the Go benchmark suite, the default time is 10s.

#### docker

`make docker` cross compiles into a linux arch64 app and then copies that over to an arch linux container by default.

## Travis

Lile generates a Travis CI yaml file which runs the test suite by default. Support is included for pushing to a docker repositoy, but needs uncommenting.

## Monitoring

Lile includes built in app monitoring, which by default is prometheus.

### Prometheus

By default Lile will add the Prometheus gRPC interceptor and collect gRPC metrics.

You can access these at `/metrics` on port `8080` by default or change them like below:

``` go
lile.NewServer(
  lile.PrometheusPort(":9999"),
  lile.PrometheusAddr("/prom"),
)
```

You can disable prometheus interception completely by setting that option:

``` go
lile.NewServer(
  lile.PrometheusEnabled(false),
)
```

## Tracing

Lile support opentracing tracers and ships with [Zipkin](http://zipkin.io/) compatiable tracing by default. Though you will need to set an ENV var to enable the feature.

#### Zipkin via HTTP

To have Lile send traces to Zipkin via the HTTP endpoint set the `ZIPKIN_HTTP_ENDPOINT` env variable

`ZIPKIN_HTTP_ENDPOINT=http://localhost:9411`

#### Zipkin via Scribe

To have Lile send traces to Zipkin using the scribe protocol set the `ZIPKIN_SCRIBE_ENDPOINT` env variable

`ZIPKIN_SCRIBE_ENDPOINT=http://localhost:9411`

#### Zipkin via Kafka

To have Lile send traces to Zipkin via Kakfa (to then be collected) set the `ZIPKIN_KAFKA_ENDPOINTS` env variable. 

It can be a single Kafka host or multiple hosts seperated by comma

`ZIPKIN_KAFKA_ENDPOINTS=10.0.0.1,10.0.0.2`

#### Opentracing

Lile supports opentracing compatible tracers via [otgrpc](https://github.com/grpc-ecosystem/grpc-opentracing/tree/master/go/otgrpc) and [opentracing-go](https://github.com/opentracing/opentracing-go)

You can set a tracer using the `Tracer` option..

``` go
zipkin := zipkin.NewTracer(
  zipkin.NewRecorder(collector, false, opts.name, opts.name)
)

lile.NewServer(
  lile.Tracer(zipkin),
)
```

You can enable/disable tracing the `TracingEnabled` option..

``` go
lile.NewServer(
  lile.TracingEnabled(false),
)
```
