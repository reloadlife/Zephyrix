
SHELL := /bin/bash
export GO111MODULE=on
export GOPROXY=https://proxy.golang.org

.DEFAULT_GOAL: all

LDFLAGS=-ldflags "-s -w "


dev: ## Run the application in dev mode
	@air -c ./.air.toml


git-tag-patch: ## Push new tag to repository with patch number incremented
	$(eval NEW_VERSION=$(shell git describe --tags --abbrev=0 | awk -F'[a-z.]' '{$$4++;print "v" $$2 "." $$3 "." $$4}'))
	@echo Version: $(NEW_VERSION)
	@git tag -a $(NEW_VERSION) -m "New Patch release"
	@git push origin $(NEW_VERSION)

git-tag-minor: ## Push new tag to repository with minor number incremented
	$(eval NEW_VERSION=$(shell git describe --tags --abbrev=0 | awk -F'[a-z.]' '{$$3++;print "v" $$2 "." $$3 "." 0}'))
	@echo Version: $(NEW_VERSION)
	@git tag -a $(NEW_VERSION) -m "New Minor release"
	@git push origin $(NEW_VERSION)

git-tag-major:  ## Push new tag to repository with major number incremented
	$(eval NEW_VERSION=$(shell git describe --tags --abbrev=0 | awk -F'[a-z.]' '{$$2++;print "v" $$2 "." 0 "." 0}'))
	@echo Version: $(NEW_VERSION)
	@git tag -a $(NEW_VERSION) -m "New Major release"
	@git push origin $(NEW_VERSION)
