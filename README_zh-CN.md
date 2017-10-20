![logo](https://raw.githubusercontent.com/lileio/lile/master/lile.png)

> **ALPHA:** Lile当前正处于Alpha版本，可能还会存在一些变化。目前，我正在收集反馈意见，并将尽快敲定Lile，以避免发生变化。

Lile是可以帮助您快速创建基于gRPC通讯，或者可以通过[gateway](https://github.com/grpc-ecosystem/grpc-gateway)创建REST通讯的发布订阅服务的一个生成器和工具集。

Lile主要是用于过创建基本结构，测试示例，Dockerfile，Makefile等基础骨架。

Lile也是一个简单的服务生成器，扩展了基本的gRPC服务器，包括诸如指标（如[Prometheus](prometheus.io)），跟踪（如[Zipkin](zipkin.io)）和发布订阅（如[Google PubSub](https://cloud.google.com/pubsub/docs/overview)等可插拔选项。

Lile在的Slack上的[Gopher Slack](https://invite.slack.golangbridge.org/) 交流渠道[#lile](https://gophers.slack.com/messages/C6RHLV3LN)

[![Build Status](https://travis-ci.org/lileio/lile.svg?branch=master)](https://travis-ci.org/lileio/lile) [![GoDoc](https://godoc.org/github.com/lileio/lile?status.svg)](https://godoc.org/github.com/lileio/lile) [![Go Report Card](https://goreportcard.com/badge/github.com/lileio/lile)](https://goreportcard.com/report/github.com/lileio/lile) [![license](https://img.shields.io/github/license/mashape/apistatus.svg)]()

[![asciicast](https://asciinema.org/a/rLAGV6nsdBreyWgXtb6bgL3Hb.png)](https://asciinema.org/a/rLAGV6nsdBreyWgXtb6bgL3Hb)

### 安装

安装Lile很容易，使用`go get`便可以安装Lile的命令行工具来生成新的服务和所需的库。

```
$ go get -u github.com/lileio/lile/...
```

您还需要安装Google的[Protocol Buffers](https://developers.google.com/protocol-buffers/)。

### 入门

通过执行`lile run`和一个短文件路径生成一个新的服务。

Lilek可以自动根据`username/service`生成一个完整的路径到`$GOPATH`下的`github.com`中。

```
$ lile new lileio/users
```

# 指南

- [安装](#安装)
- [创建服务](#创建服务)
- [服务定义](#服务定义)
- [生成RPC方法](#生成RPC方法)
- [编写并运行测试](#编写并运行测试)
- [使用生成的命令行](#使用生成的命令行)
- [自定义命令行](#自定义命令行)
- [暴露Prometheus指标](#暴露Prometheus采集指标)
- [发布和订阅](#发布和订阅)
- [发布事件](#发布事件)
- [自动发布事件](#自动发布事件)
- [订阅事件](#订阅事件)
- [追踪](#追踪)

## 安装

首先，你需要确保您在您安装Lile之前已经安装了Go。

安装Lile很容易，使用`go get`便可以安装Lile的命令行工具来生成新的服务和所需的库。

```
$ go get github.com/lileio/lile/...
```

您还需要安装Google的[Protocol Buffers][Protocol Buffers](https://developers.google.com/protocol-buffers/)。

在MacOS你可以使用`brew install protobuf`来安装。

## 创建服务

Lile使用生成器来快速生成新的Lile服务。

Lile遵循Go关于$GOPATH的约定（参见[如何写Go](https://golang.org/doc/code.html#Workspaces)），并且自动解析您的新服务的名称，以在正确的位置创建服务。

如果您的Github用户名是lileio，并且您想创建一个新的服务为了发布消息到Slack，您可以使用如下命令：

```
lile new lileio/slack
```

这将创建一个项目到`$GOPATH/src/github.com/lileio/slack`

## 服务定义

Lile服务主要使用gRPC，因此使用[protocol buffers](https://developers.google.com/protocol-buffers/)作为接口定义语言（IDL），用于描述有效负载消息的服务接口和结构。 如果需要，可以使用其他替代品。

我强烈建议您先阅读[Google API设计](https://cloud.google.com/apis/design/)文档，以获得有关RPC方法和消息的一般命名的好建议，以及如果需要，可以将其转换为REST/JSON。

您可以在Lile中发现一个简单的例子[`account_service`](https://github.com/lileio/account_service)

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

## 生成RPC方法

默认情况下，Lile将创建一个RPC方法和一个简单的请求和响应消息。

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

我们来修改一下使它能够提供真正的服务，并添加自己的方法。

我们来创建一个`Announce`方法向Slack发布消息。

我们假设Slack团队和身份验证已经由服务配置来处理，所以我们服务的用户只需要提供一个房间和他们的消息。 该服务将发送特殊的空响应，因为我们只需要知道是否发生错误，也不需要知道其他任何内容。

现在我们的`proto`文件看起来像这样：

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

现在我们运行`protoc`工具我们的文件，以及Lile生成器插件。

```
protoc -I . slack.proto --lile-server_out=. --go_out=plugins=grpc:$GOPATH/src
```

Lile提供了一个`Makefile`，每个项目都有一个已经配置的`proto`构建步骤。 所以我们可以运行它。

```
make proto
```

我们可以看到，Lile将在`server`目录中为我们创建两个文件。

```
$ make proto
protoc -I . slack.proto --lile-server_out=. --go_out=plugins=grpc:$GOPATH/src
2017/07/12 15:44:01 [Creating] server/announce.go
2017/07/12 15:44:01 [Creating test] server/announce_test.go
```

我们来看看Lile为我们创建的`announce.go`文件。

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

接下来我们实现这个生成的方法，让我们从测试开始吧！


## 编写并运行测试

当您使用Lile生成RPC方法时，也会创建一个对应的测试文件。例如，给定我们的`announce.go`文件，Lile将在同一目录中创建`announce_test.go`

看起来如下:

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

您现在可以使用`Makefile`运行测试，并运行`make test`命令

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

我们的测试失败了，因为我们还没有实现我们的方法，在我们的方法中返回一个“未实现”的错误。

让我们在`announce.go`中实现`Announce`方法，这里是一个使用`nlopes`的[slack library](https://github.com/nlopes/slack)的例子。

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

我们再次修改我们的测试用力，然后再次运行我们的测试

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
现在如果我使用我的Slack令牌作为环境变量运行测试，我应该看到通过测试！

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

## 使用生成的命令行

生成您的服务时，Lile生成一个命令行应用程序。 您可以使用自己的命令行扩展应用程序或使用内置的命令行来运行服务。

运行没有任何参数的命令行应用程序将打印生成的帮助。

例如`go run orders/main.go`

### 服务

运行`serve`将运行RPC服务。

### 订阅

运行`subscribe`订阅者将会收到你的订阅发布事件。

### up

运行`up`将同时运行RPC服务器和发布订阅的订阅者。

## 自定义命令行

要添加您自己的命令行，您可以使用[cobra](https://github.com/spf13/cobra)，它是Lile的内置的命令行生成器。

``` bash
$ cd orders
$ cobra add import
```

您现在可以编辑生成的文件，以创建您的命令行，`cobra`会自动将命令行的名称添加到帮助中。 

## 暴露Prometheus采集指标

默认情况下，Lile将[Prometheus](prometheus.io)的采集指标暴露在`:9000/metrics`。

如果您的服务正在运行，您可以使用cURL来预览Prometheus指标。

```
$ curl :9000/metrics
```

你应该看到如下的一些输出：

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

Lile的Prometheus指标实现使用go-grpc-promesheus的拦截器将其内置到自身的gPRC中，提供如以下的指标：

```
grpc_server_started_total
grpc_server_msg_received_total
grpc_server_msg_sent_total
grpc_server_handling_seconds_bucket
```

有关使用Prometheus的更多信息，收集和绘制这些指标，请参阅[Prometheus入门](https://prometheus.io/docs/introduction/getting_started/)

有关gRPC Prometheus查询的示例，请参阅[查询示例](https://github.com/grpc-ecosystem/go-grpc-prometheus#useful-query-examples)。

## 发布和订阅

虽然大多数服务将主要通过RPC进行通信，但Lile提供了一个[pubsub](https://github.com/lileio/lile/tree/master/pubsub)来执行[Publish & Subscribe](https://en.wikipedia.org/wiki/Publish%E2%80%93subscribe_pattern)通信，这在开发异步的请求时特别有用。

You can either manually publish events or use the gRPC [middleware](https://github.com/lileio/lile/blob/master/pubsub/interceptor.go) to automatically publish an event when a RPC method is called.

发布者与订阅者是松耦合的，甚至不知道它们的存在。大多数[Lile](https://github.com/lileio)已经内置了事件钩子，但是您也可以很容易的添加事件到您的服务中。您可以手动发布事件或使用gRPC[中间件](https://github.com/lileio/lile/blob/master/pubsub/interceptor.go)在调用RPC方法时自动发布事件。

Lile的发布订阅是基于每个用户存在“至少一次”消息传递。换句话说，给定一个发布事件`account_created`的`account_service`（发布者），如果`email_service`（subscriber）和`fraud_detection_service`（subscriber）的多个实例正在运行，则每个`email_service`和`fraud_detection_service`都只有一个实例收到一条消息。

## 发布事件

如果配置了发布订阅（通过环境自动或手动进行），那么只需要一个简单的“Publish”调用。

以下是“订单”服务的“Get”方法示例：

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

`Publish`需要一个`context.Context`用于跟踪和指标，一个主题名称和一个`proto.Msg`，它是可以序列化到protobuf的任何对象。

## 自动发布事件

当使用了Lile的gPRC[中间件](https://github.com/lileio/lile/blob/master/pubsub/interceptor.go)，在调用gRPC方法时，可以自动发布事件。

在我们的`main.go`中，我们可以添加发布订阅拦截器，将我们的gRPC方法映射到我们的发布订阅主题。

拦截器将自动发布gRPC响应到该主题。

``` go
lile.AddPubSubInterceptor(map[string]string{
	"Create": "account_service.created",
	"Update": "account_service.updated",
})
```

## 订阅事件

默认情况下，Lile将在项目中生成一个具有一些订阅事件基本设置的`subscribers.go`文件。

[`lile.Subscriber`](https://godoc.org/github.com/lileio/lile/pubsub#Subscriber)借口用于订阅符合`Setup`事件规范任何主题的事件。
```go
type OrdersServiceSubscriber struct{}

func (s *OrdersServiceSubscriber) Setup(c *pubsub.Client) {
	c.On("shipments.updated", s.ShipmentUpdate, 30*time.Second, true)
}

func (s *OrdersServiceSubscriber) ShipmentUpdate(sh *shipments.Shipment) {
	// do something with sh
}
```

[Handler interface](https://godoc.org/github.com/lileio/lile/pubsub#Handler)是用于监听符合发布订阅的任何内容的功能。

Protobuf消息自动解码。

## 追踪

Lile已经建立了跟踪，将[opentracing](http://opentracing.io/) 兼容的跟踪器设置为`GlobalTracer`，默认情况下，Lile报告所有gRPC方法和发布/订阅操作。

![](https://2.bp.blogspot.com/-0pFWb8zb-Cg/WPb9qKoDwDI/AAAAAAAAD2g/VjUFl1-_tYgy6zpzw0iyjfwh3gh0rg92wCLcB/s640/go-2.png)

### Zipkin

要通过HTTP将所有跟踪事件发送到[Zipkin](http://zipkin.io)，请将环境变量`ZIPKIN_SERVICE_HOST`设置为Zipkin服务的DNS名称。 如果服务已经运行命名为“zipkin”，Kubernetes将自动将`ZIPKIN_SERVICE_HOST`暴露给容器。

### Stackdriver (Google Cloud Platform) Trace

如果你想使用追踪而又不想维护Zipkin，Stackdriver提供了一个`zipkin-collector`映像，它将侦听Zipkin跟踪，转换并发送到Stackdriver。更多请参阅[Google Cloud Tracing](https://cloud.google.com/trace/docs/zipkin#option_1_using_a_container_image_to_set_up_your_server)