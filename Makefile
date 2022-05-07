# NOTE: this Makefile is meant to provide a simplified entry point for humans to
# run all of the critical steps to verify one's changes are harmonious in
# nature. Keeping target bodies to one line each and abstaining from make magic
# are very important so that maintainers and contributors can focus their
# attention on files that are primarily Go.

.PHONY: all
all: generate lint tag-test test check-bin tag-check-bin gfmrun toc

.PHONY: generate
generate:
	go run internal/build/build.go generate

.PHONY: lint
lint:
	go run internal/build/build.go vet

.PHONY: tag-test
tag-test:
	go run internal/build/build.go -tags urfave_cli_no_docs test

.PHONY: test
test:
	go run internal/build/build.go test

.PHONY: check-bin
check-bin:
	go run internal/build/build.go check-binary-size

.PHONY: tag-check-bin
tag-check-bin:
	go run internal/build/build.go -tags urfave_cli_no_docs check-binary-size

.PHONY: gfmrun
gfmrun:
	go run internal/build/build.go gfmrun docs/v2/manual.md

.PHONY: toc
toc:
	go run internal/build/build.go toc docs/v2/manual.md
