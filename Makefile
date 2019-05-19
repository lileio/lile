.PHONY: test statik
test: statik
	go test ./... -v -cover

statik:
	go get github.com/rakyll/statik
	statik -src=template
	cd protoc-gen-lile-server && statik -src=template

default: test
