.PHONY: test cover install bin test-env rm-test-env

LDFLAGS := -X main.version=$(version)

test:
ifneq ($(shell go fmt ./...),)
	$(error code not formatted. Please run 'go fmt')
endif
  @go test github.com/davrodpin/mole/... -v -race -coverprofile=coverage.txt -covermode=atomic
cover: test
	go tool cover -html=coverage.txt

install:
	@go install github.com/davrodpin/mole/cmd/mole

bin:
ifeq ($(version),)
	$(error usage: make bin version=X.Y.Z)
endif
	GOOS=darwin GOARCH=amd64 go build -o bin/mole -ldflags "$(LDFLAGS)" github.com/davrodpin/mole/cmd/mole
	cd bin && tar c mole | gzip > mole$(version).darwin-amd64.tar.gz && rm mole && cd -
	GOOS=linux GOARCH=amd64 go build -o bin/mole -ldflags "$(LDFLAGS)" github.com/davrodpin/mole/cmd/mole
	cd bin && tar c mole | gzip > mole$(version).linux-amd64.tar.gz && rm mole && cd -
	GOOS=windows GOARCH=amd64 go build -o bin/mole.exe -ldflags "$(LDFLAGS)" github.com/davrodpin/mole/cmd/mole
	cd bin && zip mole$(version).windows-amd64.zip mole.exe && rm -f mole.exe && cd -

add-network:
	-@docker network create --subnet=192.168.33.0/24 mole
rm-network:
	-@docker network rm mole

rm-mole-http:
	-@docker rm -f mole_http
mole-http: rm-mole-http
	@docker build \
		--tag mole_http:latest \
		./test-env/http-server/
	@docker run \
		--detach \
		--network mole \
		--ip 192.168.33.11 \
		--name mole_http mole_http:latest

rm-mole-ssh:
	-@docker rm -f mole_ssh
mole-ssh: rm-mole-ssh
	@docker build \
		--tag mole_ssh:latest \
		./test-env/ssh-server/
	@docker run \
		--detach \
		--network mole \
		--ip 192.168.33.10 \
		--publish 22122:22 \
		--name mole_ssh mole_ssh:latest

test-env: add-network mole-http mole-ssh

rm-test-env: rm-mole-http rm-mole-ssh rm-network

