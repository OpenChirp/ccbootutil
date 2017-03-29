# Makes the distributable binaries
# The idea was to make this runnable by anyone, even without the GOPATH setup
#
# Craig Hesling
# March 19, 2017

# Directory to place builds in
BUILDS=builds

SOURCES=$(wildcard *.go)

# Sorta need this obscure mkdir line in all targets because GNU Make seems
# to check the directory access/modification time and constantly forces a rebuild
# of the enclosed build targets.
MKDIR_LINE=mkdir -p $(BUILDS)
BUILD_LINE=GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $@ $(SOURCES)

.PHONY: all clean

# Build binary for all platforms
all: $(addprefix $(BUILDS)/, ccbootutil ccbootutil.osx ccbootutil.exe)

$(BUILDS)/ccbootutil: $(SOURCES)
$(BUILDS)/ccbootutil: GOOS=linux GOARCH=amd64
$(BUILDS)/ccbootutil: BINNAME=ccbootutil
$(BUILDS)/ccbootutil:
	$(MKDIR_LINE)
	$(BUILD_LINE)

$(BUILDS)/ccbootutil.osx: $(SOURCES)
$(BUILDS)/ccbootutil.osx: GOOS=darwin GOARCH=amd64
$(BUILDS)/ccbootutil.osx: BINNAME=ccbootutil.osx
$(BUILDS)/ccbootutil.osx:
	$(MKDIR_LINE)
	$(BUILD_LINE)

$(BUILDS)/ccbootutil.exe: $(SOURCES)
$(BUILDS)/ccbootutil.exe: GOOS=windows GOARCH=386
$(BUILDS)/ccbootutil.exe: BINNAME=ccbootutil.exe
$(BUILDS)/ccbootutil.exe:
	$(MKDIR_LINE)
	$(BUILD_LINE)

clean:
	$(RM) -r $(BUILDS)
