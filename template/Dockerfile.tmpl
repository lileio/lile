FROM golang:alpine AS build-env
WORKDIR /{{ .Name }}
RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh
COPY go.mod /{{ .Name }}/go.mod
COPY go.sum /{{ .Name }}/go.sum
RUN go mod download
COPY . /{{ .Name }}
RUN CGO_ENABLED=0 GOOS=linux go build -o build/{{.Name}} ./{{.Name}}


FROM scratch
COPY --from=build-env /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build-env /{{ .Name }}/build/{{.Name}} /
ENTRYPOINT ["/{{.Name}}"]
CMD ["up", "--grpc-port=80"]
