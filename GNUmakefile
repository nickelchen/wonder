GOTOOLS = github.com/mitchellh/gox github.com/kardianos/govendor
VERSION = $(shell awk -F\" '/^const Version/ { print $$2; exit }' cmd/wonder/version.go)
GITSHA:=$(shell git rev-parse HEAD)
GITBRANCH:=$(shell git symbolic-ref --short HEAD 2>/dev/null)

default:: test

# bin generates the releasable binaries
bin:: tools
	@sh -c "'$(CURDIR)/scripts/build.sh'"

# cov generates the coverage output
cov:: tools
	gocov test ./... | gocov-html > /tmp/coverage.html
	open /tmp/coverage.html

# dev creates binaries for testing locally - these are put into ./bin and
# $GOPATH
dev::
	@ENV_DEV=1 sh -c "'$(CURDIR)/scripts/build.sh'"

# dist creates the binaries for distibution
dist::
	@sh -c "'$(CURDIR)/scripts/dist.sh' $(VERSION)"

get-tools::
	go get -u -v $(GOTOOLS)


# test runs the test suite
test:: tools
	@go list ./... | grep -v -E '^vendor' | xargs -n1 go test $(TESTARGS)

# testrace runs the race checker
testrace::
	go test -race `govendor list -no-status +local` $(TESTARGS)

tools::
	@which gox 2>/dev/null ; if [ $$? -eq 1 ]; then \
        $(MAKE) get-tools; \
    fi

# updatedeps installs all the dependencies needed to test, build, and run
updatedeps:: tools
	govendor list -no-status +vendor | xargs -n1 go get -u
	govendor update +vendor

vet:: tools
	@echo "--> Running go tool vet $(VETARGS) ."
	@govendor list -no-status +local \
        | cut -d '/' -f 4- \
        | xargs -n1 \
            go tool vet $(VETARGS) ;\
    if [ $$? -ne 0 ]; then \
        echo ""; \
        echo "Vet found suspicious constructs. Please check the reported constructs"; \
        echo "and fix them if necessary before submitting the code for reviewal."; \
    fi

.PHONY: default bin cov dev dist get-tools test testrace tools updatedeps vet
