# NOTE: this Makefile is meant to provide a simplified entry point for humans to
# run all of the critical steps to verify one's changes are harmonious in
# nature. Keeping target bodies to one line each and abstaining from make magic
# are very important so that maintainers and contributors can focus their
# attention on files that are primarily Go.

.PHONY: all
all: generate vet tag-test test check-binary-size tag-check-binary-size gfmrun v2diff

# NOTE: this is a special catch-all rule to run any of the commands
# defined in internal/build/build.go with optional arguments passed
# via GFLAGS (global flags) and FLAGS (command-specific flags), e.g.:
#
#   $ make test GFLAGS='--packages cli'
%:
	go run internal/build/build.go $(GFLAGS) $* $(FLAGS)

.PHONY: tag-test
tag-test:
	go run internal/build/build.go -tags urfave_cli_no_docs test

.PHONY: tag-check-binary-size
tag-check-binary-size:
	go run internal/build/build.go -tags urfave_cli_no_docs check-binary-size

.PHONY: gfmrun
gfmrun:
	go run internal/build/build.go gfmrun docs/v2/manual.md

.PHONY: docs
docs:
	mkdocs build

.PHONY: docs-deps
docs-deps:
	pip install -r mkdocs-requirements.txt

.PHONY: serve-docs
serve-docs:
	mkdocs serve
