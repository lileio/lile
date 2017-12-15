![logo](https://raw.githubusercontent.com/lileio/lile/master/lile.png)

> **ALPHA:** Lile is currently considered "Alpha" in that things may change. Currently I am gathering feedback and will finalise Lile shortly to avoid breaking changes going forward.

Lile is a generator and set of tools/libraries to help you quickly create services that communicate via [gRPC](grpc.io) (REST via a [gateway](https://github.com/grpc-ecosystem/grpc-gateway)) and publish subscribe.

The primary focus of Lile is to remove the boilerplate when creating new services by creating a basic structure, test examples, Dockerfile, Makefile etc.

As well a simple service generator Lile extends the basic gRPC server to include pluggable options like metrics (e.g. [Prometheus](prometheus.io)), tracing (e.g. [Zipkin](zipkin.io)) and PubSub (e.g. [Google PubSub](https://cloud.google.com/pubsub/docs/overview)).

Chat to me on Slack on the [Gopher Slack](https://invite.slack.golangbridge.org/) channel [#lile](https://gophers.slack.com/messages/C6RHLV3LN)

[![Build Status](https://travis-ci.org/lileio/lile.svg?branch=master)](https://travis-ci.org/lileio/lile) [![GoDoc](https://godoc.org/github.com/lileio/lile?status.svg)](https://godoc.org/github.com/lileio/lile) [![Go Report Card](https://goreportcard.com/badge/github.com/lileio/lile)](https://goreportcard.com/report/github.com/lileio/lile) [![license](https://img.shields.io/github/license/mashape/apistatus.svg)]()

[![asciicast](https://asciinema.org/a/rLAGV6nsdBreyWgXtb6bgL3Hb.png)](https://asciinema.org/a/rLAGV6nsdBreyWgXtb6bgL3Hb)

### Installation

Installing Lile is easy, using `go get` you can install the cmd line app to generate new services and the required libraries.

```
$ go get -u github.com/lileio/lile/...
```

You will also need Google's [Protocol Buffers](https://developers.google.com/protocol-buffers/) installed.

### Getting Started

To generate a new service, run `lile new` with a short folder path.

Lile is smart enough to evaluate `username/service` to a full `$GOPATH` directory and defaults to `github.com`.

```
$ lile new lileio/users
```

# Guide

- [Installation](#installation)
- [Creating a Service](#creating-a-service)
- [Service Definition](#service-definitions)
- [Generating RPC Methods](#generating-rpc-methods)
- [Running and Writing Tests](#running--writing-tests)
- [Using the Generated cmds](#using-the-generated-cmds)
- [Adding your own cmds](#adding-your-own-cmds)
- [Exposing Prometheus Metrics](#exposing--collecting-prometheus-metrics)
- [Publish & Subscribe](#publish--subscribe)
- [Publishing an Event](#publishing-an-event)
- [Automatically Publishing Events](#automatically-publishing-events)
- [Subscribing to Events](#subscribing-to-events)
- [Tracing](#tracing)

## Installation

First, you need to have a working Go installation, once you have Go installed you can then install Lile.

Installing Lile is easy, using `go get` you can install the cmd line app to generate new services and the required libraries.

```
$ go get github.com/lileio/lile/...
```

You will also need Google's [Protocol Buffers](https://developers.google.com/protocol-buffers/) installed.

On MacOS you can simply `brew install protobuf`

## Creating a Service

Lile comes with a 'generator' to quickly generate new Lile services.

Lile follows Go's conventions around `$GOPATH` (see [How to Write Go](https://golang.org/doc/code.html#Workspaces)) and is smart enough to parse your new service's name to create the service in the right place.

If your Github username was `lileio` and you wanted to create a new service for posting to Slack you might use the following command.

```
lile new lileio/slack
```

This will create a project in `$GOPATH/src/github.com/lileio/slack`

## Service Definitions

Lile services mainly speak gRPC and therefore uses [protocol buffers](https://developers.google.com/protocol-buffers/) as the Interface Definition Language (IDL) for describing both the service interface and the structure of the payload messages. It is possible to use other alternatives if desired.

I highly recommend reading the [Google API Design](https://cloud.google.com/apis/design/) docs for good advice around general naming of RPC methods and messages and how they might translate to REST/JSON if needed.

An example of a service definition can be found in the Lile [`account_service`](https://github.com/lileio/account_service)

``` protobuf
service AccountService {
  rpc List (ListAccountsRequest) returns (ListAccountsResponse) {}
  rpc GetById (GetByIdRequest) returns (Account) {}
  rpc GetByEmail (GetByEmailRequest) returns (Account) {}
  rpc AuthenticateByEmail (AuthenticateByEmailRequest) returns (Account) {}
  rpc GeneratePasswordToken (GeneratePasswordTokenRequest) returns (GeneratePasswordTokenResponse) {}
  rpc ResetPassword (ResetPasswordRequest) returns (Account) {}
  rpc ConfirmAccount (ConfirmAccountRequest) returns (Account) {}
  rpc Create (CreateAccountRequest) returns (Account) {}
  rpc Update (UpdateAccountRequest) returns (Account) {}
  rpc Delete (DeleteAccountRequest) returns (google.protobuf.Empty) {}
}
```

## Generating RPC Methods

By default Lile will create a example RPC method and a simple message for request and response.

``` protobuf
syntax = "proto3";
option go_package = "github.com/lileio/slack";
package slack;

message Request {
  string id = 1;
}

message Response {
  string id = 1;
}

service Slack {
  rpc Read (Request) returns (Response) {}
}
```

Let's modify this to be a real service and add our own method.

We're going to create an `Announce` method that will announce a message to a Slack room.

We're assuming that the Slack team and authentication is already handled by the services configuration, so a user of our service only needs to provide a `room` and their `message`. The service is going to send the special `Empty` response, since we only need to know if an error occurred and don't need to know anything else.

Our `proto` file now looks like this...

``` protobuf
syntax = "proto3";
option go_package = "github.com/lileio/slack";
import "google/protobuf/empty.proto";
package slack;

message AnnounceRequest {
  string channel = 1;
  string msg = 2;
}

service Slack {
  rpc Announce (AnnounceRequest) returns (google.protobuf.Empty) {}
}
```

We now run the `protoc` tool with our file and the Lile method generator plugin.

```
protoc -I . slack.proto --lile-server_out=. --go_out=plugins=grpc:$GOPATH/src
```

Handily, Lile provides a `Makefile` with each project that has a `proto` build step already configured. So we can just run that.

```
make proto
```

We can see that Lile will create two files for us in the `server` directory.

```
$ make proto
protoc -I . slack.proto --lile-server_out=. --go_out=plugins=grpc:$GOPATH/src
2017/07/12 15:44:01 [Creating] server/announce.go
2017/07/12 15:44:01 [Creating test] server/announce_test.go
```

Let's take a look at the `announce.go` file that's created for us.

``` go
package server

import (
    "errors"

    "github.com/golang/protobuf/ptypes/empty"
    "github.com/lileio/slack"
    context "golang.org/x/net/context"
)

func (s SlackServer) Announce(ctx context.Context, r *slack.AnnounceRequest) (*empty.Empty, error) {
  return nil, errors.New("not yet implemented")
}
```

We can now fill in this generated method with the correct implementation. But let's start with a test!

## Running & Writing Tests

When you generate an RPC method with Lile a counterpart test file is also created. For example, given our `announce.go` file, Lile will create `announce_test.go` in the same directory.

This should look something like the following..

``` go
package server

import (
	"testing"

	"github.com/lileio/slack"
	"github.com/stretchr/testify/assert"
	context "golang.org/x/net/context"
)

func TestAnnounce(t *testing.T) {
	ctx := context.Background()
	req := &slack.AnnounceRequest{}

	res, err := cli.Announce(ctx, req)
	assert.Nil(t, err)
	assert.NotNil(t, res)
}

```

You can now run the tests using the `Makefile` and running `make test`...

```
$ make test
=== RUN   TestAnnounce
--- FAIL: TestAnnounce (0.00s)
        Error Trace:    announce_test.go:16
        Error:          Expected nil, but got: &status.statusError{Code:2, Message:"not yet implemented", Details:[]*any.Any(nil)}
        Error Trace:    announce_test.go:17
        Error:          Expected value not to be nil.
FAIL
coverage: 100.0% of statements
FAIL    github.com/lileio/slack/server  0.011s
make: *** [test] Error 2

```

Our test failed because we haven't implemented our method, at the moment we're returning an error of "unimplemented" in our method.

Let's implement the `Announce` method in `announce.go`, here's an example using `nlopes`' [slack library](https://github.com/nlopes/slack).

``` go
package server

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/lileio/slack"
	sl "github.com/nlopes/slack"
	context "golang.org/x/net/context"
)

var api = sl.New(os.Getenv("SLACK_TOKEN"))

func (s SlackServer) Announce(ctx context.Context, r *slack.AnnounceRequest) (*empty.Empty, error) {
	_, _, err := api.PostMessage(r.Channel, r.Msg, sl.PostMessageParameters{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &empty.Empty{}, nil
}
```

Let's fill out our testing request and then run our tests again...

``` go
package server

import (
	"testing"

	"github.com/lileio/slack"
	"github.com/stretchr/testify/assert"
	context "golang.org/x/net/context"
)

func TestAnnounce(t *testing.T) {
	ctx := context.Background()
	req := &slack.AnnounceRequest{
		Channel: "@alex",
		Msg:     "hellooo",
	}

	res, err := cli.Announce(ctx, req)
	assert.Nil(t, err)
	assert.NotNil(t, res)
}
```
Now if I run the tests with my Slack token as an `ENV` variable, I should see a passing test!

```
$ alex@slack: SLACK_TOKEN=zbxkkausdkasugdk make test
go test -v ./... -cover
?       github.com/lileio/slack [no test files]
=== RUN   TestAnnounce
--- PASS: TestAnnounce (0.32s)
PASS
coverage: 75.0% of statements
ok      github.com/lileio/slack/server  0.331s  coverage: 75.0% of statements
?       github.com/lileio/slack/slack   [no test files]
?       github.com/lileio/slack/slack/cmd       [no test files]
?       github.com/lileio/slack/subscribers     [no test files]
```

## Using the Generated cmds

Lile generates a cmd line application when you generate your service. You can extend the app with your own cmds or use the built-in cmds to run the service.

Running the cmd line app without any arguments will print the generated help.

For example `go run orders/main.go`

### serve
Running `serve` will run the RPC server.

### subscribe
Running `subscribe` will listen to pubsub events with your subscribers.

### up
Running `up` will run both the RPC server and the pubsub subscribers.

## Adding your own cmds

To add your own cmd, you can use the built in generator from [cobra](https://github.com/spf13/cobra) which powers Lile's cmds

``` bash
$ cd orders
$ cobra add import
```

You can now edit the file generated to create your cmd, `cobra` will automatically add the cmd's name to the help. 

## Exposing & Collecting Prometheus Metrics

By default Lile collects [Prometheus](prometheus.io) metrics and exposes them at `:9000/metrics`. You can set a custom port by setting the env `PROMETHEUS_PORT`.  

If your service is running, you can use cURL to preview the Prometheus metrics

```
$ curl :9000/metrics
```

You should see something along the lines of...

```
# HELP go_gc_duration_seconds A summary of the GC invocation durations.
# TYPE go_gc_duration_seconds summary
go_gc_duration_seconds{quantile="0"} 0
go_gc_duration_seconds{quantile="0.25"} 0
go_gc_duration_seconds{quantile="0.5"} 0
go_gc_duration_seconds{quantile="0.75"} 0
go_gc_duration_seconds{quantile="1"} 0
go_gc_duration_seconds_sum 0
go_gc_duration_seconds_count 0
...
...
```

The Lile Prometheus metrics implementation plugs itself into gPRC using an interceptor using [go-grpc-prometheus](https://github.com/grpc-ecosystem/go-grpc-prometheus) providing metrics such as;

```
grpc_server_started_total
grpc_server_msg_received_total
grpc_server_msg_sent_total
grpc_server_handling_seconds_bucket
```

For more on using Prometheus, collecting and graphing these metrics, see [Getting started](https://prometheus.io/docs/introduction/getting_started/) at Prometheus.io

And see [useful query examples](https://github.com/grpc-ecosystem/go-grpc-prometheus#useful-query-examples) for examples of useful gRPC Prometheus queries.

## Publish & Subscribe

Whilst most services will communicate predominantly via RPC, Lile provides a [library](https://github.com/lileio/pubsub) for doing [Publish & Subscribe](https://en.wikipedia.org/wiki/Publish%E2%80%93subscribe_pattern) (or Pub Sub) communication.

This is particularly helpful when developing a service that needs to be updated or do some work when another service has performed an action, but you don't want to hold up the request.

Publishers are loosely coupled to subscribers, and need not even know of their existence. Most [Lile services](https://github.com/lileio) will already provide events you can hook into, but you can easily add events to your own service. 

Lile's Pub Sub is based on "at least once" delivery of message **per subscriber**. In other words, given an `account_service` (publisher) that publishes the event `account_created`, if multiple instances of an `email_service` (subscriber) and `fraud_detection_service` (subscriber) are running, only one instance of each `email_service` and `fraud_detection_service` will each receive a message.

## Publishing an Event

(PubSub has moved to it's own package located at https://github.com/lileio/pubsub)

If Lile pubsub is configured (which happens via env vars automatically or manually) then only a simple call to `Publish` is required.

Here's an example `Get` method on an "orders" service.

``` go
func (s OrdersServer) Get(ctx context.Context, r *orders.GetRequest) (*orders.GetResponse, error) {
	o, err := getOrder(r.Id)
	if err != nil {
		return nil, err
	}

	res := &orders.GetResponse{
		Id:   o.Id,
		Name: o.Name,
	}
	pubsub.Publish(ctx, "orders_service.Get", res)
	return res, nil
}
```

`Publish` takes a `context.Context` for tracing and metrics, a topic name and a `proto.Msg`, which is any object that can be serialised to protobuf.

## Automatically Publishing Events

Automatically publishing events has been removed, due to their misuse in general and confusion. Now you should manually publish events.

## Subscribing to Events

Lile generates projects by default with a `subscribers.go` file with some basic setup to subscribe to events.

Subscribers conform to the [`lile.Subscriber`](https://godoc.org/github.com/lileio/pubsub#Subscriber) interface which has a special `Setup` event for subscribing to events from any topic.

```go
type OrdersServiceSubscriber struct{}

func (s *OrdersServiceSubscriber) Setup(c *pubsub.Client) {
	c.On("shipments.updated", s.ShipmentUpdate, 30*time.Second, true)
}

func (s *OrdersServiceSubscriber) ShipmentUpdate(sh *shipments.Shipment) {
	// do something with sh
}
```

Functions that listen to topics can take anything that conforms to pubsub's [Handler interface](https://godoc.org/github.com/lileio/pubsub#Handler)

Protobuf messages are automatically decoded.

## Tracing

Lile has built in tracing that reports to [opentracing](http://opentracing.io/) compatible tracers set to the `GlobalTracer` and by default, Lile with report all gRPC methods and pubsub publish/subscribing actions.

![](https://2.bp.blogspot.com/-0pFWb8zb-Cg/WPb9qKoDwDI/AAAAAAAAD2g/VjUFl1-_tYgy6zpzw0iyjfwh3gh0rg92wCLcB/s640/go-2.png)

### Zipkin

To have Lile send all tracing events to [Zipkin](http://zipkin.io) via HTTP, set the `ZIPKIN_SERVICE_HOST` ENV variable to the DNS name of your Zipkin service. Kubernetes will expose the `ZIPKIN_SERVICE_HOST` automatically to a container if there is service already running named `zipkin`.


### Stackdriver (Google Cloud Platform) Trace

Stackdriver provide a `zipkin-collector` image that will listen for Zipkin traces, convert and send them to Stackdriver. It's quite awesome if you're looking for tracing but don't want to maintain Zipkin! See the [Google Cloud Tracing](https://cloud.google.com/trace/docs/zipkin#option_1_using_a_container_image_to_set_up_your_server) docs for more
