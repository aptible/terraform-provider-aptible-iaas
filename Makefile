HOSTNAME = aptible.com
NAMESPACE = aptible
NAME = aptible-iaas
BINARY = terraform-provider-${NAME}
VERSION = 0.0.0+local
GOOS ?= darwin
GOARCH ?= amd64
TARGET ?= ${GOOS}_${GOARCH}

default: install

build:
	# https://www.terraform.io/plugin/debugging#debugging-caveats
	GOOS=${GOOS} GOARCH=${GOARCH} go build -o -gcflags="all=-N -l" -o ./bin/${BINARY}_${VERSION}_${TARGET}

clean:
	rm -rf "$$HOME/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}" || true
.PHONY: clean

install: clean build
	mkdir -p "$$HOME/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${TARGET}"
	cp ./bin/${BINARY}_${VERSION}_${TARGET} "$$HOME/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${TARGET}"
	@echo "Installed as provider aptible.com/aptible/aptible-iaas version 0.0.0+local"

test:
	cd test && go test -v || exit 1
.PHONY: test

testacc:
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m

resource:
	cp ./internal/provider/asset/aws/acm/resources.go ./internal/provider/asset/aws/ecs_web/resources.go
	sed -i '' 's/package acm/package ecsweb/g' ./internal/provider/asset/aws/ecs_web/resources.go

	cp ./internal/provider/asset/aws/acm/resources.go ./internal/provider/asset/aws/rds/resources.go
	sed -i '' 's/package acm/package rds/g' ./internal/provider/asset/aws/rds/resources.go

	cp ./internal/provider/asset/aws/acm/resources.go ./internal/provider/asset/aws/redis/resources.go
	sed -i '' 's/package acm/package redis/g' ./internal/provider/asset/aws/redis/resources.go

	cp ./internal/provider/asset/aws/acm/resources.go ./internal/provider/asset/aws/vpc/resources.go
	sed -i '' 's/package acm/package vpc/g' ./internal/provider/asset/aws/vpc/resources.go

	cp ./internal/provider/asset/aws/acm/resources.go ./internal/provider/asset/aws/secret/resources.go
	sed -i '' 's/package acm/package secret/g' ./internal/provider/asset/aws/secret/resources.go

	cp ./internal/provider/asset/aws/acm/resources.go ./internal/provider/asset/aws/ecs_compute/resources.go
	sed -i '' 's/package acm/package ecscompute/g' ./internal/provider/asset/aws/ecs_compute/resources.go
.PHONY: resource

debug:
	dlv debug --accept-multiclient --continue --headless --listen=0.0.0.0:33000 ./main.go -- -debug
.PHONY: debug

dc:
	dlv connect 0.0.0.0:33000
.PHONY: debug-connect

release:
	GOOS=darwin GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_darwin_amd64
	GOOS=freebsd GOARCH=386 go build -o ./bin/${BINARY}_${VERSION}_freebsd_386
	GOOS=freebsd GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_freebsd_amd64
	GOOS=freebsd GOARCH=arm go build -o ./bin/${BINARY}_${VERSION}_freebsd_arm
	GOOS=linux GOARCH=386 go build -o ./bin/${BINARY}_${VERSION}_linux_386
	GOOS=linux GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_linux_amd64
	GOOS=linux GOARCH=arm go build -o ./bin/${BINARY}_${VERSION}_linux_arm
	GOOS=openbsd GOARCH=386 go build -o ./bin/${BINARY}_${VERSION}_openbsd_386
	GOOS=openbsd GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_openbsd_amd64
	GOOS=solaris GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_solaris_amd64
	GOOS=windows GOARCH=386 go build -o ./bin/${BINARY}_${VERSION}_windows_386
	GOOS=windows GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_windows_amd64
