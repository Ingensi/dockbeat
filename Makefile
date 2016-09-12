BEATNAME=dockbeat
BEAT_DIR=github.com/ingensi
SYSTEM_TESTS=false
TEST_ENVIRONMENT=false
ES_BEATS=./vendor/github.com/elastic/beats
GOPACKAGES=$(shell glide novendor)
GOFILES_NOVENDOR = $(shell find . -type f -name '*.go' -not -path "./vendor/*")
PREFIX?=.

# Path to the libbeat Makefile
-include $(ES_BEATS)/libbeat/scripts/Makefile

.PHONY: init
init:
	glide update --strip-vcs
	make update
	git init

.PHONY: commit
commit:
	git add README.md CONTRIBUTING.md
	git commit -m "Initial commit"
	git add LICENSE
	git commit -m "Add the LICENSE"
	git add .gitignore .gitattributes
	git commit -m "Add git settings"
	git add .
	git reset -- .travis.yml
	git commit -m "Add dockbeat"
	git add .travis.yml
	git commit -m "Add Travis CI"

.PHONY: update-deps
update-deps:
	glide update --strip-vcs

# Checks project and source code if everything is according to standard
.PHONY: check
check:
	@gofmt -l ${GOFILES_NOVENDOR} | read && echo "Code differs from gofmt's style" 1>&2 && exit 1 || true
	go vet ${GOPACKAGES}

# Run integration tests. Unit tests are run as part of the integration tests. It runs all tests with race detection enabled.
.PHONY: integration-tests
integration-tests: prepare-tests
	$(GOPATH)/bin/gotestcover -race -coverprofile=${COVERAGE_DIR}/integration.cov -covermode=atomic ${GOPACKAGES}

# Generates a coverage report from the existing coverage files
# It assumes that some covrage reports already exists, otherwise it will fail
.PHONY: coverage-report
coverage-report:
	python ${ES_BEATS}/scripts/aggregate_coverage.py -o ./${COVERAGE_DIR}/full.cov ./${COVERAGE_DIR}
	go tool cover -html=./${COVERAGE_DIR}/full.cov -o ${COVERAGE_DIR}/full.html


# This is called by the beats packer before building starts
.PHONY: before-build
before-build:
