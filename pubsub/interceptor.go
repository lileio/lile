package pubsub

import (
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"

	"golang.org/x/net/context"

	"google.golang.org/grpc"
)

// UnaryServerInterceptor is a gRPC server-side interceptor that automatically publishes events
func UnaryServerInterceptor(c *Client, methods map[string]string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		resp, err := handler(ctx, req)
		if err != nil {
			return resp, err
		}

		if methods == nil {
			return resp, err
		}

		methodParts := strings.Split(info.FullMethod, "/")
		method := methodParts[len(methodParts)-1]
		intercept, exist := methods[method]
		if exist {
			msg, ok := (resp).(proto.Message)
			if !ok {
				logrus.Errorf("Couldn't convert interface{} into proto.Message %v", resp)
			}

			go func() {
				err := c.Provider.Publish(ctx, intercept, msg)
				if err != nil {
					logrus.Errorf("Couldn't publish to topic:%s, err:%s", intercept, err)
					return
				}

				logrus.Infof("Published msg for method:%s in topic:%s", method, intercept)
			}()
		}

		return resp, err
	}
}
