.PHONY: build
build:
	GO111MODULE=on go build ./cmd/kube-better-node

.PHONY: test
test:
	go test -v -race -buildvcs ./...

.PHONY: clean
clean:
	rm -f kube-better-node
