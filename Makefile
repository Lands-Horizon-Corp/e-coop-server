.PHONY: \
	cache-clean security-enforce-blocklist \
	db-reset db-migrate db-seed \
	server run-debug test build build-debug \
	actiongraph actiongraph-why profiler resurrect teleport webdev deploy build-all-trace

########################################
# Go helpers
########################################

cache-clean:
	go clean -cache

security-enforce-blocklist:
	@echo "Enforcing security blocklist..."
	go run . security-enforce-blocklist

########################################
# Database
########################################

db-reset:
	go run . db-reset

db-migrate:
	go run . db-migrate

db-seed:
	go run . db-seed

########################################
# Run / Dev
########################################

server:
	@echo "Starting server with Air (live reload)..."
	air

run-debug:
	@echo "Running server with debug flags..."
	go run -gcflags='all=-N -l' . server

test:
	@echo "Running tests..."
	go test -v ./services/horizon_test

########################################
# Build
########################################

build:
	@echo "Building app..."
	go build -o app .

build-debug:
	@echo "Building app with debug flags..."
	go build -gcflags='all=-N -l' -o app .

########################################
# Build + Dependency Analysis (Actiongraph)
########################################

actiongraph: cache-clean
	@bash -c '\
	start=$$(date +%s); \
	echo "=== Actiongraph: FULL PROJECT ==="; \
	echo "WARNING: slow and memory-heavy"; \
	go build -debug-actiongraph=compile.json ./...; \
	end=$$(date +%s); \
	echo "Actiongraph finished in $$((end - start)) seconds"; \
	echo "Output: compile.json"; \
	'

actiongraph-why:
	@bash -c '\
	echo "Rendering compile dependency WHY-graph"; \
	actiongraph graph --why ./... -f compile.json > compile-why.dot; \
	dot -Tsvg -Grankdir=LR < compile-why.dot > compile-why.svg; \
	echo "Output: compile-why.svg"; \
	'

########################################
# Profiling / Diagnostics
########################################

profiler:
	@bash -c '\
	start=$$(date +%s); \
	echo "=== PROFILER MODE ==="; \
	echo "WARNING: very slow, high memory usage"; \
	go clean -cache -modcache -testcache -fuzzcache; \
	go run -gcflags="all=-N -l" . server; \
	end=$$(date +%s); \
	echo "Profiler run finished in $$((end - start)) seconds"; \
	'

########################################
# Utility / Recovery
########################################

resurrect:
	@bash -c '\
	start=$$(date +%s); \
	echo "=== RESURRECT ==="; \
	go clean -cache -modcache -testcache -fuzzcache; \
	git pull; \
	go run . db-reset; \
	go run . db-migrate; \
	go run . db-seed; \
	go run . server; \
	end=$$(date +%s); \
	echo "Total resurrect time: $$((end - start)) seconds"; \
	'

teleport:
	clear
	git pull
	go run . server

webdev:
	code .

########################################
# Deployment
########################################

deploy:
	@echo "Deploying to Fly.io..."
	fly deploy
	fly logs

########################################
# Full build trace
########################################

build-all-trace: cache-clean
	@bash -c '\
	mkdir -p tmp; \
	echo "=== FULL BUILD WITH DEBUG TRACE ==="; \
	start=$$(date +%s); \
	go build -debug-trace=tmp/trace.json ./...; \
	end=$$(date +%s); \
	echo "Full build + trace finished in $$((end - start)) seconds"; \
	echo "Trace file saved to tmp/trace.json"; \
	'
