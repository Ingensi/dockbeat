GODEP=$(GOPATH)/bin/godep
PREFIX?=/build

GOFILES = $(shell find . -type f -name '*.go')
dockerbeat: $(GOFILES)
	# first make sure we have godep
	go get github.com/tools/godep
	$(GODEP) go build

.PHONY: getdeps
getdeps:
	go get -t -u -f

.PHONY: test
test:
	$(GODEP) go test ./...


.PHONY: install_cfg
install_cfg:
	cp etc/dockerbeat.yml $(PREFIX)/dockerbeat-linux.yml
	cp etc/dockerbeat.template.json $(PREFIX)/dockerbeat.template.json
	# darwin
	cp etc/dockerbeat.yml $(PREFIX)/dockerbeat-darwin.yml
	# win
	cp etc/dockerbeat.yml $(PREFIX)/dockerbeat-win.yml

.PHONY: cover
cover:
	# gotestcover is needed to fetch coverage for multiple packages
	go get github.com/pierrre/gotestcover
	GOPATH=$(shell $(GODEP) path):$(GOPATH) $(GOPATH)/bin/gotestcover -coverprofile=profile.cov -covermode=count github.com/ingensi/dockerbeat/...
	mkdir -p cover
	$(GODEP) go tool cover -html=profile.cov -o cover/coverage.html

.PHONY: clean
clean:
	rm dockerbeat || true
	-rm -r cover
