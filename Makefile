.PHONY: clean build wake test refresh resurrect clense teleport webdev

# Binary name
BINARY := ecoop

# Go helpers
clean:
	go clean -cache
	rm -f $(BINARY)

build:
	go build -o $(BINARY)

wake:
	./$(BINARY) server

test:
	go test -v "./services/horizon_tes"

refresh:
	./$(BINARY) db refresh

# Combined / utility targets
clense:
	go clean -cache -modcache -testcache -fuzzcache
	rm -f $(BINARY)

resurrect: clense
	clear
	git pull
	$(MAKE) build
	$(MAKE) refresh
	$(MAKE) wake

teleport:
	clear
	git pull
	$(MAKE) wake

webdev:
	code & googit pulle-chrome
