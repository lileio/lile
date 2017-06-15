# ![logo](https://raw.githubusercontent.com/lileio/lile/docs/lile.png)

> **ALPHA:** Lile is currently considered "Alpha" in that things may change. Currently I am gathering feedback and will finalise Lile shortly to avoid breaking changes going foward.

Lile is a generator and set of tools/libraries to help you quickly create services  that communicate via [gRPC](grpc.io) (REST via a [gateway](https://github.com/grpc-ecosystem/grpc-gateway)) and publish subscribe.

The primary focus of Lile is to remove the boilerplate when creating new services by creating a basic structure, test examples, Dockerfile, Makefile etc.

As well a simple service generator Lile extends the basic gRPC server to include pluggable options like metrics (e.g. [Prometheus](prometheus.io)), tracing (e.g. [Zipkin](zipkin.io)) and PubSub (e.g. [Google PubSub](https://cloud.google.com/pubsub/docs/overview)).

[![Build Status](https://travis-ci.org/lileio/lile.svg?branch=master)](https://travis-ci.org/lileio/lile) [![GoDoc](https://godoc.org/github.com/lileio/lile?status.svg)](https://godoc.org/github.com/lileio/lile) [![Go Report Card](https://goreportcard.com/badge/github.com/lileio/lile)](https://goreportcard.com/report/github.com/lileio/lile) [![license](https://img.shields.io/github/license/mashape/apistatus.svg)]()

![](https://dl.dropboxusercontent.com/s/z91on1e6x2k9gvj/2017-06-15%2012.04.45.gif?dl=0)

### Installation

Installing Lile is easy, using `go get` you can install the cmd line app to generate new services and the required libaries.

```
$ go get -u github.com/lileio/lile/lile
```

### Getting Started

To generate a new service, run `lile new` with a short folder path. 

Lile is smart enough to evaluate `username/service` to a full `$GOPATH` directory and defaults to `github.com`.

```
$ lile new lileio/users
```