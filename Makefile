TEST? = $$(go list ./... | grep -v 'vendor')
HOSTNAME = aptible.com
NAMESPACE = aptible
NAME = aptible-iaas
BINARY = terraform-provider-${NAME}
VERSION = 0.0.0+local
TARGET = darwin_amd64
LOCAL_ARCH ?= amd64
LOCAL_TARGET ?= darwin_${LOCAL_ARCH}

default: install

build:
	go build -o ${BINARY}

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


build-local:
	GOOS=darwin GOARCH=${LOCAL_ARCH} go build -o ./bin/${BINARY}_${VERSION}_${LOCAL_TARGET}

local-install: build-local
	# delete existing if it's already been saved
	rm -rf "$$HOME/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}" || true
	mkdir -p "$$HOME/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/$(LOCAL_TARGET)"
	cp ./bin/${BINARY}_${VERSION}_${LOCAL_TARGET} "$$HOME/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/$(LOCAL_TARGET)"
	@echo "Installed as provider aptible.com/aptible/aptible-iaas version 0.0.0+local"

install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

test:
	go test -i $(TEST) || exit 1
	echo $(TEST) | xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4

testacc:
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m

resource:
	cp ./internal/provider/assets/aws/acm/resources.go ./internal/provider/assets/aws/ecs/resources.go
	sed -i 's/package acm/package ecs/g' ./internal/provider/assets/aws/ecs/resources.go

	cp ./internal/provider/assets/aws/acm/resources.go ./internal/provider/assets/aws/rds/resources.go
	sed -i 's/package acm/package rds/g' ./internal/provider/assets/aws/rds/resources.go

	cp ./internal/provider/assets/aws/acm/resources.go ./internal/provider/assets/aws/redis/resources.go
	sed -i 's/package acm/package redis/g' ./internal/provider/assets/aws/redis/resources.go

	cp ./internal/provider/assets/aws/acm/resources.go ./internal/provider/assets/aws/vpc/resources.go
	sed -i 's/package acm/package vpc/g' ./internal/provider/assets/aws/vpc/resources.go

	cp ./internal/provider/assets/aws/acm/resources.go ./internal/provider/assets/aws/secret/resources.go
	sed -i 's/package acm/package secret/g' ./internal/provider/assets/aws/secret/resources.go
.PHONY: resource
