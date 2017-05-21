.PHONY: test
test:
	go test ./... -v -cover

default: test
