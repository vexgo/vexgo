.PHONY: build test clean run verify frontend-build

frontend-build:
	cd admin-frontend && pnpm install && pnpm build
	rm -rf internal/admin/dist && cp -r admin-frontend/dist internal/admin/dist

build: frontend-build
	go build -o go-cms ./cmd/server/

test:
	go test ./... -v

clean:
	rm -f go-cms

run: build
	./go-cms &

verify: run
	sleep 2
	curl -s http://localhost:8080/ping
	kill %1 2>/dev/null || true

.DEFAULT_GOAL := build
