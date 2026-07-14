# go/Makefile — the Go SDK lane of the kion-sdk monorepo.
#
# This lane consumes the shared, language-agnostic spec at
# $(SPEC_DIR)/<v>/openapi3.json (produced by ../preprocess, driven by the root
# Makefile) and generates + builds the typed Go client.
#
# Dev targets (build/test/lint/fmt/vet/ci) operate on this Go module in place —
# run them from this directory. Generation targets are normally invoked by the
# root Makefile's orchestrator (`make gen LANG=go VERSION=<v>`), but work
# standalone too as long as $(SPEC_DIR)/<v>/openapi3.json exists.

.PHONY: help build build-smoketest test test-integration lint fmt vet ci \
        ci-fmt ci-vet ci-lint ci-test install-ogen scaffold-version clean

.DEFAULT_GOAL := help

# Shared spec directory, produced by ../preprocess (root-level `spec/`).
SPEC_DIR ?= ../spec

# Pinned ogen version. MUST match the ogen version required in this lane's
# go.mod so locally regenerated code and the committed client stay consistent.
OGEN_VERSION := v1.20.1

## help: Show this help.
help:
	@echo "kion-sdk-go lane — typed Go SDK for the Kion API"
	@echo ""
	@echo "  make build                    go build ./..."
	@echo "  make test                     unit tests"
	@echo "  make test-integration         integration tests (KION_URL/KION_API_KEY, build tag)"
	@echo "  make build-smoketest          build the smoketest binary (build tag)"
	@echo "  make ci                        fmt-check, vet, lint, race tests (mirrors CI)"
	@echo "  make install-ogen              install pinned ogen ($(OGEN_VERSION))"
	@echo "  make generate-<v>              regenerate one version from \$$(SPEC_DIR)/<v>/openapi3.json"
	@echo "  make scaffold-version VERSION=<v>   create hand-written client.go/auth.go for a new version"
	@echo ""
	@echo "  SPEC_DIR=$(SPEC_DIR)"

## install-ogen: Install the pinned ogen generator.
install-ogen:
	go install github.com/ogen-go/ogen/cmd/ogen@$(OGEN_VERSION)

# Regenerate the typed client for ONE version from the shared spec.
# We rm oas_*.go explicitly (instead of ogen's -clean) so the hand-written
# client.go and auth.go in the same package survive. gofmt normalizes ogen's
# alignment so committed and CI-regenerated code are byte-identical.
# Usage: make generate-v3_16   (or via root: make gen LANG=go VERSION=v3_16)
generate-%:
	rm -f generated/$*/oas_*.go
	ogen -target ./generated/$* -package $* $(SPEC_DIR)/$*/openapi3.json
	gofmt -s -w generated/$*/

## build: Compile the root package and every generated sub-package.
build:
	go build ./...

## build-smoketest: Build the smoketest demo binary (behind the smoketest tag).
build-smoketest:
	go build -tags smoketest -o smoketest ./cmd/smoketest

## test: Run unit tests (integration suite is behind the integration tag).
test:
	go test -v -count=1 ./...

## test-integration: Run integration tests against a live Kion instance.
test-integration:
	go test -v -count=1 -tags integration ./...

## lint: Run golangci-lint.
lint:
	golangci-lint run

## fmt: Format all Go source files.
fmt:
	gofmt -s -w .

## vet: Run go vet.
vet:
	go vet ./...

## ci: Run all checks (mirrors the GitLab pipeline's quality + test stages).
ci: ci-fmt ci-vet ci-lint ci-test

ci-fmt:
	@test -z "$$(gofmt -l .)" || { echo "Files need gofmt:"; gofmt -l .; exit 1; }

ci-vet:
	go vet ./...

ci-lint:
	golangci-lint run

ci-test:
	go test -race -count=1 -v -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out | tail -1

# scaffold-version: create the hand-written wrapper files (client.go, auth.go)
# for a new version by copying generated/master/ with sed substitutions. Run
# after the version's ogen output exists. Refuses to overwrite the master
# template. Usage: make scaffold-version VERSION=v3_17
scaffold-version:
	@set -e; \
	[ -n "$(VERSION)" ] || { echo "ERROR: specify VERSION=<version>"; exit 1; }; \
	if [ "$(VERSION)" = "master" ]; then \
		echo "ERROR: master is the template; cannot scaffold it from itself"; exit 1; \
	fi; \
	if [ ! -f generated/master/client.go ] || [ ! -f generated/master/auth.go ]; then \
		echo "ERROR: generated/master/client.go or auth.go missing."; exit 1; \
	fi; \
	if [ ! -f generated/$(VERSION)/oas_client_gen.go ]; then \
		echo "ERROR: generated/$(VERSION)/oas_client_gen.go missing. Generate the version first."; exit 1; \
	fi; \
	label=$$(echo "$(VERSION)" | sed 's/^v3_/3./'); \
	sed -e "s|^package master|package $(VERSION)|" \
	    -e "s|master (unreleased)|$$label|g" \
	    -e "s|generated/master|generated/$(VERSION)|g" \
	    -e "s|master\\.New|$(VERSION).New|g" \
	    generated/master/client.go > generated/$(VERSION)/client.go; \
	sed -e "s|^package master|package $(VERSION)|" \
	    generated/master/auth.go > generated/$(VERSION)/auth.go; \
	echo "=> scaffolded generated/$(VERSION)/{client.go,auth.go} from master template"

## clean: Remove build artifacts (leaves generated/*/oas_*.go in place).
clean:
	rm -f smoketest coverage.out
