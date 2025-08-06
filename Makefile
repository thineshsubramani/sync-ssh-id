OUTDIR := bin
PKG := ./cmd/sync-ssh-id
LDFLAGS := -s -w

PLATFORMS := \
	linux/amd64 \
	linux/arm64 \
	windows/amd64 \
	windows/arm64 \
	darwin/amd64 \
	darwin/arm64

all: clean build-all

build-all:
	@mkdir -p $(OUTDIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$${platform%/*}; \
		GOARCH=$${platform##*/}; \
		BIN=$(OUTDIR)/sync-ssh-id_$${GOOS}_$${GOARCH}; \
		if [ "$$GOOS" = "windows" ]; then BIN=$$BIN.exe; fi; \
		echo "Building $$BIN..."; \
		CGO_ENABLED=0 GOOS=$$GOOS GOARCH=$$GOARCH go build -o $$BIN -ldflags "$(LDFLAGS)" $(PKG); \
	done

clean:
	rm -rf $(OUTDIR)

# Single target build example:
# make build GOOS=linux GOARCH=amd64
build:
	@mkdir -p $(OUTDIR)
	@BIN=$(OUTDIR)/sync-ssh-id_$(GOOS)_$(GOARCH); \
	if [ "$(GOOS)" = "windows" ]; then BIN=$$BIN.exe; fi; \
	echo "Building $$BIN..."; \
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $$BIN -ldflags "$(LDFLAGS)" $(PKG)
