GO = go
GOFMT = gofmt

.PHONY: check test

check:
	@echo -n "Go version: "
	@$(GO) version
	@echo "Doing gofmt"
	@$(GOFMT) -l -d .
	@test $$($(GOFMT) -l .) && exit 1 ; echo -n
	@echo "Doing go vet"
	@$(GO) vet ./... || exit 1

test:
	@$(GO) test -parallel 4 -v -run ^Test -failfast

cover:
	@$(GO) test -parallel 4 -v -run ^Test -failfast -coverprofile cover.cover
	@$(GO) tool cover -html=cover.cover
