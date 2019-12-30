GOFMT=goimports

goimports:
	${GOFMT} -w ./

goimports-check:
	$(eval diff_files = $(shell ${GOFMT} -l ./))
	@if [ -n "${diff_files}" ]; then \
		echo "Please run 'make goimports' to fix the format errors in the files:"; \
		echo "\t${diff_files}"; \
		exit 1; \
	fi;
