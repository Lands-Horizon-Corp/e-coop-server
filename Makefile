.PHONY: \ 
	profiler

########################################
# Unified Profiler / Actiongraph
########################################

########################################
# Unified Profiler / Actiongraph
########################################

profiler:
	@bash -c '\
	start=$$(date +%s); \
	echo "=== PROFILER & ACTIONGRAPH MODE ==="; \
	echo "WARNING: very slow, high memory usage"; \
	echo "Output directory: /tmp/profiler-output"; \
	mkdir -p /tmp/profiler-output; \
	\
	echo ""; \
	echo "Step 1: Clearing ALL caches"; \
	go clean -cache -modcache -testcache -fuzzcache; \
	\
	echo ""; \
	echo "Step 2: Building with debug-actiongraph in current folder"; \
	go build -debug-actiongraph=/tmp/profiler-output/compile.json ./...; \
	echo "Actiongraph build finished: /tmp/profiler-output/compile.json"; \
	\
	echo ""; \
	echo "Step 3: Rendering compile dependency WHY-graph"; \
	actiongraph graph --why ./... -f /tmp/profiler-output/compile.json > /tmp/profiler-output/compile-why.dot; \
	dot -Tsvg -Grankdir=LR < /tmp/profiler-output/compile-why.dot > /tmp/profiler-output/compile-why.svg; \
	echo "WHY-graph SVG generated: /tmp/profiler-output/compile-why.svg"; \
	\
	echo ""; \
	echo "Step 4: Running server with profiler/debug flags"; \
	go run -gcflags="all=-N -l" . server; \
	\
	end=$$(date +%s); \
	echo ""; \
	echo "=== PROFILER & ACTIONGRAPH COMPLETE ==="; \
	echo "Total time: $$((end - start)) seconds"; \
	'
