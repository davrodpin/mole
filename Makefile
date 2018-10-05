.PHONY: bin

.install:
	@go install github.com/davrodpin/mole/cmd/mole
.bin:
ifeq ($(version),)
	$(error usage: make bin version=X.Y.Z)
endif
	GOOS=darwin GOARCH=amd64 go build -o bin/mole github.com/davrodpin/mole/cmd/mole
	tar c bin/mole | gzip > bin/mole$(version).darwin-amd64.tar.gz
	rm bin/mole
	GOOS=linux GOARCH=amd64 go build -o bin/mole github.com/davrodpin/mole/cmd/mole
	tar c bin/mole | gzip > bin/mole$(version).linux-amd64.tar.gz
	rm bin/mole
test:
	@go test ./... -race -coverprofile=coverage.txt -covermode=atomic

