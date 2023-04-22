REPO_ROOT = $(shell git rev-parse --show-toplevel)

# BINDIR ?= /usr/local/bin
# BINDIT must be in PATH
BINDIR ?= $(REPO_ROOT)/bin

INSTALL = $(QUIET)install
GO_BUILD = CGO_ENABLED=0 go build

# TARGET - name of the target binary
TARGET ?= helper

$(TARGET): clean
	$(GO_BUILD) -o $(TARGET) ./cmd/main.go

install: $(TARGET)
	$(INSTALL) -m 0755 -d $(BINDIR)
	$(INSTALL) -m 0755 $(TARGET) $(BINDIR)

clean:
	rm -rf $(BINDIR)/$(TARGET)
