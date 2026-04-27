MODULES := . redis bbolt memcached

# ──────────────────────────────────────────────
# help
# ──────────────────────────────────────────────
.PHONY: help
help:
	@echo "gormcache monorepo targets"
	@echo ""
	@echo "Tagging (VERSION required):"
	@echo "  make tag-core       VERSION=v0.0.17   tag core module"
	@echo "  make tag-redis      VERSION=v0.1.1    tag redis plugin  (prefix: redis/)"
	@echo "  make tag-bbolt      VERSION=v0.1.1    tag bbolt plugin  (prefix: bbolt/)"
	@echo "  make tag-memcached  VERSION=v0.1.1    tag memcached plugin"
	@echo "  make tag-plugins    VERSION=v0.1.1    tag all three plugins with same version"
	@echo "  make push-tags                        push all local tags to origin"
	@echo ""
	@echo "Dependency update:"
	@echo "  make update-core    VERSION=v0.0.17   go get core@VERSION in each plugin"
	@echo "  make tidy                             go mod tidy in all modules"
	@echo ""
	@echo "Verification:"
	@echo "  make build                            go build ./... in all modules"
	@echo "  make vet                              go vet ./... in all modules"
	@echo "  make test                             go test ./... in all modules"
	@echo "  make versions                         show latest tag per module"

# ──────────────────────────────────────────────
# tagging
# ──────────────────────────────────────────────
.PHONY: tag-core tag-redis tag-bbolt tag-memcached tag-plugins push-tags

_require-version:
	@test -n "$(VERSION)" || (echo "ERROR: VERSION is required. Example: make $(MAKECMDGOALS) VERSION=v0.1.1"; exit 1)

tag-core: _require-version
	git tag -a $(VERSION) -m "core $(VERSION)"
	@echo "Tagged: $(VERSION)"

tag-redis: _require-version
	git tag -a redis/$(VERSION) -m "redis plugin $(VERSION)"
	@echo "Tagged: redis/$(VERSION)"

tag-bbolt: _require-version
	git tag -a bbolt/$(VERSION) -m "bbolt plugin $(VERSION)"
	@echo "Tagged: bbolt/$(VERSION)"

tag-memcached: _require-version
	git tag -a memcached/$(VERSION) -m "memcached plugin $(VERSION)"
	@echo "Tagged: memcached/$(VERSION)"

tag-plugins: _require-version
	git tag -a redis/$(VERSION)     -m "redis plugin $(VERSION)"
	git tag -a bbolt/$(VERSION)     -m "bbolt plugin $(VERSION)"
	git tag -a memcached/$(VERSION) -m "memcached plugin $(VERSION)"
	@echo "Tagged: redis/$(VERSION)  bbolt/$(VERSION)  memcached/$(VERSION)"

push-tags:
	git push origin --tags

# ──────────────────────────────────────────────
# dependency management
# ──────────────────────────────────────────────
.PHONY: update-core tidy

update-core: _require-version
	cd redis     && go get github.com/rgglez/gormcache@$(VERSION)
	cd bbolt     && go get github.com/rgglez/gormcache@$(VERSION)
	cd memcached && go get github.com/rgglez/gormcache@$(VERSION)

tidy:
	@for m in $(MODULES); do \
		echo "==> tidy $$m"; \
		(cd $$m && go mod tidy); \
	done

# ──────────────────────────────────────────────
# verification
# ──────────────────────────────────────────────
.PHONY: build vet test

build:
	@for m in $(MODULES); do \
		echo "==> build $$m"; \
		(cd $$m && go build ./...); \
	done

vet:
	@for m in $(MODULES); do \
		echo "==> vet $$m"; \
		(cd $$m && go vet ./...); \
	done

test:
	@for m in $(MODULES); do \
		echo "==> test $$m"; \
		(cd $$m && go test ./...); \
	done

# ──────────────────────────────────────────────
# introspection
# ──────────────────────────────────────────────
.PHONY: versions

versions:
	@echo "core:      $$(git tag --sort=-v:refname | grep -E '^v[0-9]' | head -1)"
	@echo "redis:     $$(git tag --sort=-v:refname | grep -E '^redis/'  | head -1)"
	@echo "bbolt:     $$(git tag --sort=-v:refname | grep -E '^bbolt/'  | head -1)"
	@echo "memcached: $$(git tag --sort=-v:refname | grep -E '^memcached/' | head -1)"
