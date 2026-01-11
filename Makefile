.PHONY: \ 
	profiler

########################################
# Unified Profiler / Actiongraph
########################################

########################################
# Unified Profiler / Actiongraph
########################################

########################################
# Unified Profiler / Actiongraph
########################################


profiler:
	@bash -c '\
	start=$$(date +%s); \
	TIMESTAMP=$$(date "+%Y-%m-%d_%H-%M-%S"); \
	OUTPUT_DIR="./tmp/$$TIMESTAMP"; \
	echo "=== PROFILER & ACTIONGRAPH & TRACE MODE ==="; \
	echo "WARNING: very slow, high memory usage"; \
	echo "Output directory: $$OUTPUT_DIR"; \
	mkdir -p $$OUTPUT_DIR; \
	\
	echo ""; \
	echo "Step 1: Clearing ALL caches"; \
	go clean -cache -modcache -testcache -fuzzcache; \
	\
	echo ""; \
	echo "Step 2: Building with debug-actiongraph"; \
	go build -debug-actiongraph=$$OUTPUT_DIR/compile.json ./...; \
	echo "Actiongraph build finished: $$OUTPUT_DIR/compile.json"; \
	\
	echo ""; \
	echo "Step 3: Building with debug-trace"; \
	go build -debug-trace=$$OUTPUT_DIR/trace.json ./...; \
	echo "Trace build finished: $$OUTPUT_DIR/trace.json"; \
	\
	echo ""; \
	echo "Step 4: Rendering compile dependency WHY-graph"; \
	actiongraph graph --why ./... -f $$OUTPUT_DIR/compile.json > $$OUTPUT_DIR/compile-why.dot; \
	dot -Tsvg -Grankdir=LR < $$OUTPUT_DIR/compile-why.dot > $$OUTPUT_DIR/compile-why.svg; \
	echo "WHY-graph SVG generated: $$OUTPUT_DIR/compile-why.svg"; \
	\
	echo ""; \
	echo "Step 5: Running server with profiler/debug flags"; \
	go run -gcflags="all=-N -l" . server; \
	\
	end=$$(date +%s); \
	echo ""; \
	echo "=== PROFILER, ACTIONGRAPH & TRACE COMPLETE ==="; \
	echo "Total time: $$((end - start)) seconds"; \
	'
