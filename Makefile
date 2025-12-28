.PHONY: \
	clean clense \
	wake run run-debug \
	test refresh \
	build build-debug \
	actiongraph-why \
	profiler resurrect teleport webdev

########################################
# Go helpers
########################################

clean:
	go clean -cache

clense:
	go clean -cache -modcache -testcache -fuzzcache

########################################
# Run / Dev
########################################

run:
	go run . server

wake: run

run-debug:
	go run -gcflags='all=-N -l' . server

test:
	go test -v ./services/horizon_test

refresh:
	go run . db refresh

########################################
# Build
########################################

build:
	go build -o ecoop-server .

build-debug:
	go build -gcflags='all=-N -l' -o ecoop-server .

########################################
# Build + Dependency Analysis (Actiongraph)
########################################

actiongraph: clense
	@bash -c '\
	start=$$(date +%s); \
	echo "=== Actiongraph: FULL PROJECT ==="; \
	echo "WARNING: slow and memory-heavy"; \
	echo "Step 1: Building with debug-actiongraph"; \
	go build -debug-actiongraph=compile.json ./...; \
	end=$$(date +%s); \
	echo "Actiongraph finished in $$((end - start)) seconds"; \
	echo "Output: compile.json"; \
	'

	'
actiongraph-why:
	@bash -c '\
	echo "Rendering compile dependency WHY-graph for $(PKG)"; \
	actiongraph graph --why $(PKG) -f compile.json > compile-why.dot; \
	dot -Tsvg -Grankdir=LR < compile-why.dot > compile-why.svg; \
	echo "Output: compile-why.svg"; \
	'

########################################
# Profiling / Diagnostics
########################################

# Full rebuild + run with profiling enabled
profiler:
	@bash -c '\
	start=$$(date +%s); \
	echo "=== PROFILER MODE ==="; \
	echo "WARNING: very slow, high memory usage"; \
	echo "Step 1: Clearing ALL caches"; \
	go clean -cache -modcache -testcache -fuzzcache; \
	echo "Step 2: Running server with debug flags"; \
	go run -gcflags='all=-N -l' . server; \
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
	echo "Step 1: Clearing all caches"; \
	go clean -cache -modcache -testcache -fuzzcache; \
	echo "Step 2: Pulling latest code"; \
	git pull; \
	echo "Step 3: Refreshing DB"; \
	go run . db refresh; \
	echo "Step 4: Starting server"; \
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
