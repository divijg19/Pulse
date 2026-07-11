.PHONY: test vet lint race tidy fmt build cert audit-render \
        audit-todo audit-deadcode doc-freshness arch-cert

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
	@for doc in README.md ARCHITECTURE.md RENDERING.md \
		internal/tui/README.md internal/tui/STATE_OWNERSHIP.md \
		internal/tui/COMPARE_CONSTITUTION.md internal/tui/COMPARE_WORKFLOW.md; do \
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

quick: test lint
	@echo "[quick] All quick gates PASS"

arch-cert: audit-render audit-todo audit-deadcode doc-freshness
	@echo "[arch] All architectural gates PASS"

cert: compilation quick race arch-cert
	@echo ""
	@echo "=========================================="
	@echo "  ALL CERTIFICATION GATES PASS"
	@echo "=========================================="
