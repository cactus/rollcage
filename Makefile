
CURDIR            ?= ${.CURDIR}
BUILDDIR          := ${CURDIR}
GOPATH            := ${BUILDDIR}:${BUILDDIR}/vendor
ITERATION         ?= 1

ROLLCAGE_VER      != git describe --always --dirty --tags|sed 's/^v//'
VERSION_VAR       := main.Version
GOTEST_FLAGS      := -cpu=1,2
GOBUILD_DEPFLAGS  :=
GOBUILD_LDFLAGS   ?=
GOBUILD_FLAGS     := ${GOBUILD_DEPFLAGS} -ldflags "${GOBUILD_LDFLAGS} -X ${VERSION_VAR}=${ROLLCAGE_VER}"
GO                := env GOPATH="${GOPATH}" go
GB                := gb
XCOMPILE_ARCHES   := darwin.amd64 freebsd.amd64 linux.amd64

.PHONY: help clean build build-setup test cover man all

help:
	@echo "Available targets:"
	@echo "  help        this help"
	@echo "  clean       clean up"
	@echo "  all         build binaries and man pages"
	@echo "  test        run tests"
	@echo "  cover       run tests with cover output"
	@echo "  build       build all binaries"
	@echo "  man         build all man pages"

clean:
	@rm -rf "${BUILDDIR}/bin"
	@rm -rf "${BUILDDIR}/pkg"
	@rm -rf "${BUILDDIR}/tar"

build-setup:
	@echo "Restoring deps..."
#	@mkdir -p "${TARBUILDDIR}"
#	@env GOPATH="${GOPATH}" gb vendor restore

build: build-setup
	@echo "Building rollcage..."
	@${GB} build ${GOBUILD_FLAGS} ...

test: build-setup
	@echo "Running tests..."
	@${GB} test ${GOTEST_FLAGS} ...

cover: build-setup
	@echo "Running tests with coverage..."
	@${GB} test -cover ${GOTEST_FLAGS} ...

${BUILDDIR}/man/%: man/%.mdoc
	@mkdir -p "${BUILDDIR}/man"
	@cat $< | sed "s#.Os ROLLCAGE VERSION#.Os ROLLCAGE ${ROLLCAGE_VER}#" > $@

#man: $(patsubst man/%.mdoc,${BUILDDIR}/man/%,$(wildcard man/*.1.mdoc))

all: build man
