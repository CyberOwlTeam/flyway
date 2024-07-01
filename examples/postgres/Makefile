ROOT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
GOBIN= $(GOPATH)/bin

define go_install
    go install $(1)
endef

$(GOBIN)/golangci-lint:
	$(call go_install,github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.2)

$(GOBIN)/gotestsum:
	$(call go_install,gotest.tools/gotestsum@latest)

.PHONY: install
install: $(GOBIN)/golangci-lint $(GOBIN)/gotestsum

.PHONY: clean
clean:
	rm $(GOBIN)/golangci-lint
	rm $(GOBIN)/gotestsum

.PHONY: lint
lint: $(GOBIN)/golangci-lint
	golangci-lint run --out-format=github-actions --path-prefix=. --verbose -c $(ROOT_DIR)/.golangci.yml --fix

.PHONY: test
test: $(GOBIN)/gotestsum
	@echo "Running $* tests..."
	gotestsum \
		--format short-verbose \
		--rerun-fails=5 \
		--packages="./..." \
		--junitfile ./build/test_unit.xml \
		-- \
		-coverprofile=./build/test_coverage.out \
		-timeout=30m

all: install lint test
