default: test

deps:
	go get golang.org/x/tools/cmd/goimports || true
	go get github.com/urfave/gfmrun/... || true
	go list ./... \
		| xargs go list -f '{{ join .Deps "\n" }}{{ printf "\n" }}{{ join .TestImports "\n" }}' \
		| grep -v github.com/urfave/cli \
		| xargs go get
	@if [ ! -f node_modules/.bin/markdown-toc ]; then \
		npm install markdown-toc ; \
	fi

gen: deps
	./runtests gen

vet:
	./runtests vet

gfmrun:
	./runtests gfmrun

v1-to-v2:
	./cli-v1-to-v2 --selftest

migrations:
	./runtests migrations

toc:
	./runtests toc

test: deps
	./runtests test

all: gen vet test gfmrun v1-to-v2 migrations toc

.PHONY: default gen vet test gfmrun migrations toc v1-to-v2 deps all
