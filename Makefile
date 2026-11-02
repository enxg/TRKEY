.PHONY: import build

import:
	@go run ./cmd/importcldr

build:
	@go run ./cmd/trkey

