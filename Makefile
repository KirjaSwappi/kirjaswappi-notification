.PHONY: format lint check

format-all:
	gofmt -w .
	goimports -w .

check-format:
	gofmt -l .
	goimports -l .

lint:
	golangci-lint run

spotless: format-all lint
	@echo "*** All codes are formatted and linted properly. ***"

check: check-format lint
	@echo "*** All checks passed! ***"

