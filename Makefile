LOCAL := $(PWD)/.local
export PATH := $(LOCAL)/bin:$(PATH)
export GOBIN := $(LOCAL)/bin

ifeq ($(OS),Windows_NT)
	BINSUFFIX:=.exe
else
	BINSUFFIX:=
endif

ifeq ($(shell which golangci-lint),)
LINTER := $(GOBIN)/golangci-lint$(BINSUFFIX)
else
LINTER := $(shell which golangci-lint)
endif

ifeq ($(shell which goreleaser),)
GORELEASER := $(GOBIN)/goreleaser$(BINSUFFIX)
else
GORELEASER := $(shell which goreleaser)
endif

$(LINTER):
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.52.2

$(GORELEASER):
	go install github.com/goreleaser/goreleaser@v1.17.2

lint: $(LINTER)
	$(LINTER) run
.PHONY: lint

snapshot: $(GORELEASER) web/admin-ui/dist
	$(GORELEASER) release --snapshot --clean
	docker tag ghcr.io/reddec/$(notdir $(CURDIR)):$$(jq -r .version dist/metadata.json)-amd64 ghcr.io/reddec/$(notdir $(CURDIR)):1
	docker tag ghcr.io/reddec/$(notdir $(CURDIR)):$$(jq -r .version dist/metadata.json)-amd64 ghcr.io/reddec/$(notdir $(CURDIR)):snapshot

local: $(GORELEASER)
	$(GORELEASER) release -f .goreleaser.local.yaml --clean

test:
	go test -v ./...

web/admin-ui/dist:
	cd web/admin-ui && npm run build

gen:
	go generate ./...

.PHONY: test