.PHONY: test statik
test: statik
	go test ./... -v -count 1 -p 1 -cover

statik:
	GO111MODULE=off go get github.com/rakyll/statik
	statik -src=template
	cd protoc-gen-lile-server && statik -src=template

default: test
