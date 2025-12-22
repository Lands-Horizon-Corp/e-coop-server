.PHONY: clean wake test refresh resurrect clense teleport webdev

# Go helpers
clean:
	go clean -cache

wake:
	go run . server

test:
	go test -v "./services/horizon_tes"

refresh:
	go run . db refresh

# Combined / utility targets
clense:
	go clean -cache -modcache -testcache -fuzzcache

resurrect:
	clear
	go clean -cache -modcache -testcache -fuzzcache
	git pull
	go run . db refresh
	go run . server

teleport:
	clear
	git pull
	go run . server

webdev:
	code & googit pulle-chrome
