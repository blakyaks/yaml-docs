VERSION="0.1.0"

yaml-docs:
	go build --ldflags="-X 'main.version=$(VERSION)'" github.com/blakyaks/yaml-docs/cmd/yaml-docs

install:
	go install github.com/blakyaks/yaml-docs/cmd/yaml-docs

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: test
test:
	go test -v ./...

.PHONY: clean
clean:
	rm -f yaml-docs

.PHONY: dist
dist:
	goreleaser release --rm-dist --snapshot --skip=sign
