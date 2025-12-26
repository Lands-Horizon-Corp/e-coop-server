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

.PHONY: resurrect

.PHONY: resurrect

.PHONY: resurrect

resurrect:
	@bash -c '\
	start=$$(date +%s); \
	echo "Starting resurrect..."; \
	echo "Step 1: Clearing caches..."; \
	go clean -cache -modcache -testcache -fuzzcache; \
	echo "Step 2: Pulling latest code..."; \
	git pull; \
	echo "Step 3: Refreshing DB..."; \
	go run . db refresh; \
	echo "Step 4: Starting server..."; \
	go run . server; \
	end=$$(date +%s); \
	echo "Total resurrect time: $$((end - start)) seconds"; \
	'



teleport:
	clear
	git pull
	go run . server

webdev:
	code & googit pulle-chrome

build:
	go build -gcflags='all=-N -l' -o app ./...

