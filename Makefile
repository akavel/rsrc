GO ?= go

.PHONY: build
build: clean
	@$(GO) build -v -ldflags="-s -w"

.PHONY: upgrade
upgrade:
	@$(GO) get -u -t ./... && go mod tidy -v

test:
	@$(GO) test ./... && echo -e "\n==>\033[32m Ok\033[m\n" || exit 1

clean:
	@rm -f *.exe *.syso
