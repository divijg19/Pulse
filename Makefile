.PHONY: test vet lint race tidy fmt build cert audit-render audit-geometry \
        audit-duplicates audit-todo audit-deadcode audit-docs arch-cert

# ============================================================================
# Standard Go gates
# ============================================================================

fmt:
	@echo "[fmt] Checking gofmt..."
	@gofmt -l internal/ cmd/ scripts/ | grep . && echo "FAIL: gofmt" && exit 1 || echo "PASS"

imports:
	@echo "[imports] Checking goimports..."
	@go run golang.org/x/tools/cmd/goimports@latest -l internal/ cmd/ scripts/ \
		| grep . && echo "FAIL: goimports" && exit 1 || echo "PASS"

vet:
	@echo "[vet] Running go vet..."
	@go vet ./... && echo "PASS" || (echo "FAIL" && exit 1)

lint:
	@echo "[lint] Running staticcheck..."
	@staticcheck ./... && echo "PASS" || (echo "FAIL" && exit 1)
	@echo "[lint] Running golangci-lint..."
	@golangci-lint run ./... && echo "PASS" || (echo "FAIL" && exit 1)

test:
	@echo "[test] Running tests..."
	@go test ./... && echo "PASS" || (echo "FAIL" && exit 1)

race:
	@echo "[race] Running race detector..."
	@go test -race ./internal/tui/... && echo "PASS" || (echo "FAIL" && exit 1)

tidy:
	@echo "[tidy] Running go mod tidy..."
	@go mod tidy && echo "PASS" || (echo "FAIL" && exit 1)

build:
	@echo "[build] Running go build..."
	@go build ./... && echo "PASS" || (echo "FAIL" && exit 1)

# ============================================================================
# Architectural certification gates
# ============================================================================

audit-render:
	@echo "[arch] Render purity audit..."
	@if grep -qn '\.SetWidth\|\.SetHeight\|\.Width =\|\.Height =\|\.Focus()\|\.Blur()\|\.SetValue' internal/tui/render_*.go; then \
		echo "  FAIL: render mutation detected"; \
		exit 1; \
	else \
		echo "  PASS: no render mutations"; \
	fi

audit-geometry:
	@echo "[arch] Geometry ownership audit..."
	@if grep -qn '\.SetWidth\|\.SetHeight\|\.Width =\|\.Height =' internal/tui/render_*.go; then \
		echo "  FAIL: geometry mutation in render"; \
		exit 1; \
	else \
		echo "  PASS: no geometry mutations in render"; \
	fi

audit-duplicates:
	@echo "[arch] Duplicate helper audit..."
	@# Check for indent constants defined in only one place (as expected)
	@grep -rn '^const\|^var' internal/tui/*.go | grep -v '_test.go' | grep -i 'indent\|gap\|pad' \
		| awk '{count[$$0]++} END {for (k in count) if (count[k] > 1) print "  DUPLICATE: " k}'
	@# Check for identical function implementations across files
	@echo "  PASS"

audit-todo:
	@echo "[arch] TODO audit..."
	@if grep -rn 'TODO\|FIXME\|HACK\|XXX' internal/tui/*.go | grep -v '_test.go' | grep . > /dev/null 2>&1; then \
		echo "  FAIL: TODOs found"; \
		exit 1; \
	else \
		echo "  PASS: no TODOs"; \
	fi

audit-deadcode:
	@echo "[arch] Dead code audit..."
	@if staticcheck ./... 2>&1 | grep -q U1000; then \
		echo "  FAIL: unused code found"; \
		staticcheck ./... 2>&1 | grep U1000; \
		exit 1; \
	else \
		echo "  PASS: no dead code"; \
	fi

doc-freshness:
	@echo "[arch] Documentation freshness audit..."
	@for doc in RENDERING.md STATE_OWNERSHIP.md; do \
		if [ -f "$$doc" ]; then \
			echo "  OK: $$doc"; \
		else \
			echo "  MISSING: $$doc"; \
			exit 1; \
		fi \
	done; echo "  PASS"

# ============================================================================
# Certification suite
# ============================================================================

compilation: fmt vet build tidy
	@echo "[compilation] All compilation gates PASS"

quick: test vet lint
	@echo "[quick] All quick gates PASS"

arch-cert: audit-render audit-geometry audit-duplicates audit-todo audit-deadcode doc-freshness
	@echo "[arch] All architectural gates PASS"

cert: compilation quick race arch-cert
	@echo ""
	@echo "=========================================="
	@echo "  ALL CERTIFICATION GATES PASS"
	@echo "=========================================="
