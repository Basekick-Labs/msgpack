test:
	go test ./...
	go test ./... -short -race -count=1 -timeout=5m
	go test ./... -run=NONE -bench=. -benchmem
	env GOOS=linux GOARCH=386 go vet ./...
	# 386 binaries run natively on amd64 hosts (CI); skipped elsewhere.
	@if [ "$$(go env GOHOSTOS)/$$(go env GOHOSTARCH)" = "linux/amd64" ]; then \
		echo "env GOARCH=386 go test ./..."; env GOARCH=386 go test ./...; \
	else \
		echo "skipping 32-bit test run (needs linux/amd64 host)"; \
	fi
	go vet
