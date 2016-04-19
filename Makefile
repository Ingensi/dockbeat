BEATNAME=dockerbeat
BEAT_DIR=github.com/ingensi
SYSTEM_TESTS=false
TEST_ENVIRONMENT=false
ES_BEATS=./vendor/github.com/elastic/beats
GOPACKAGES=$(shell glide novendor)
PREFIX?=.

# Path to the libbeat Makefile
-include $(ES_BEATS)/libbeat/scripts/Makefile

.PHONY: init
init:
	glide update  --no-recursive
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
	git commit -m "Add dockerbeat"
	git add .travis.yml
	git commit -m "Add Travis CI"

.PHONY: update-deps
update-deps:
	glide update --no-recursive --strip-vcs

.PHONY: fullupdate
fullupdate:
	$(MAKE) update
	bash ./scripts/fullupdate.sh ${BEATNAME} ${BEAT_DIR}/${BEATNAME}

# This is called by the beats packer before building starts
.PHONY: before-build
before-build:
