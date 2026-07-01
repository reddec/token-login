LOCAL := $(PWD)/.local
export PATH := $(LOCAL)/bin:$(PATH)
export GOBIN := $(LOCAL)/bin

ifeq ($(OS),Windows_NT)
	BINSUFFIX:=.exe
else
	BINSUFFIX:=
endif


lint:
	golangci-lint run
.PHONY: lint

snapshot: web/admin-ui/dist
	goreleaser release --snapshot --clean
	docker tag ghcr.io/reddec/$(notdir $(CURDIR)):$$(jq -r .version dist/metadata.json)-amd64 ghcr.io/reddec/$(notdir $(CURDIR)):1
	docker tag ghcr.io/reddec/$(notdir $(CURDIR)):$$(jq -r .version dist/metadata.json)-amd64 ghcr.io/reddec/$(notdir $(CURDIR)):snapshot

local:
	goreleaser release -f .goreleaser.local.yaml --clean

test:
	go test -v ./...

web/admin-ui/dist:
	cd web/admin-ui && npm ci && npm run build

gen:
	go generate ./...


run: web/admin-ui/dist
	go run ./cmd/token-login/main.go

.PHONY: test run