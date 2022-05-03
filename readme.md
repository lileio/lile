![logo](./docs/logo.png)
--

![Actions Status](https://github.com/lileio/lile/workflows/Test/badge.svg) [![](https://godoc.org/github.com/lileio/lile?status.svg)](http://godoc.org/github.com/lileio/lile)

Lile is a application generator (think `create-react-app`, `rails new` or `django startproject`) for gRPC services in Go and a set of tools/libraries.

The primary focus of Lile is to remove the boilerplate when creating new services by creating a basic structure, test examples, Dockerfile, Makefile etc.

Lile comes with basic pre setup with pluggable options for things like...

* Metrics (e.g. [Prometheus](https://prometheus.io))
* Tracing (e.g. [Zipkin](https://zipkin.io))
* PubSub (e.g. [Google PubSub](https://cloud.google.com/pubsub/docs/overview))
* Service Discovery

### Installation

Installing Lile is easy, using `go get` you can install the cmd line app to generate new services and the required libraries. First you'll need Google's [Protocol Buffers](https://developers.google.com/protocol-buffers/) installed.


```
$ brew install protobuf
$ go get -u github.com/lileio/lile/...
```

# Guide

- [Creating a Service](#creating-a-service)
- [Service Definition](#service-definitions)
- [Generating RPC Methods](#generating-rpc-methods)
- [Running and Writing Tests](#running--writing-tests)
- [Using the Generated cmds](#using-the-generated-cmds)
- [Adding your own cmds](#adding-your-own-cmds)

## Creating a Service

Lile comes with a 'generator' to quickly generate new Lile services.

Lile follows Go's conventions around `$GOPATH` (see [How to Write Go](https://golang.org/doc/code.html#Workspaces)) and is smart enough to parse your new service's name to create the service in the right place.

If your Github username was `tessthedog` and you wanted to create a new service for posting to Slack you might use the following command.

```
lile new --name tessthedog/slack
```

Follow the command line instructions and this will create a new project folder for you with everything you need to continue.

## Service Definitions

Lile creates [gRPC](https://grpc.io/) and therefore uses [protocol buffers](https://developers.google.com/protocol-buffers/) as the language for describing the service methods, the requests and responses.

I highly recommend reading the [Google API Design](https://cloud.google.com/apis/design/) docs for good advice around general naming of RPC methods and messages and how they might translate to REST/JSON, via the [gRPC gateway](https://github.com/grpc-ecosystem/grpc-gateway)

An example of a service definition can be found in the Lile example project [`account_service`](https://github.com/lileio/account_service)

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
option go_package = "github.com/tessthedog/slack";
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
option go_package = "github.com/tessthedog/slack";
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
    "github.com/tessthedog/slack"
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

	"github.com/tessthedog/slack"
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
FAIL    github.com/tessthedog/slack/server  0.011s
make: *** [test] Error 2

```

Our test failed because we haven't implemented our method, at the moment we're returning an error of "unimplemented" in our method.

Let's implement the `Announce` method in `announce.go`, here's an example using `nlopes`' [slack library](https://github.com/nlopes/slack).

``` go
package server

import (
	"os"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/tessthedog/slack"
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

	"github.com/tessthedog/slack"
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
?       github.com/tessthedog/slack [no test files]
=== RUN   TestAnnounce
--- PASS: TestAnnounce (0.32s)
PASS
coverage: 75.0% of statements
ok      github.com/tessthedog/slack/server  0.331s  coverage: 75.0% of statements
?       github.com/tessthedog/slack/slack   [no test files]
?       github.com/tessthedog/slack/slack/cmd       [no test files]
?       github.com/tessthedog/slack/subscribers     [no test files]
```

## Using the Generated cmds

Lile generates a cmd line application based on [cobra](https://github.com/spf13/cobra) when you generate your service. You can extend the app with your own cmds or use the built-in cmds to run the service.

Running the cmd line app without any arguments will print the generated help.

For example `go run orders/main.go`

### up
Running `up` will run both the RPC server and the pubsub subscribers.

```go run orders/main.go up```

## Adding your own cmds

To add your own cmd, you can use the built in generator from [cobra](https://github.com/spf13/cobra) which powers Lile's cmds

``` bash
$ cd orders
$ cobra add import
```

You can now edit the file generated to create your cmd, `cobra` will automatically add the cmd's name to the help.
