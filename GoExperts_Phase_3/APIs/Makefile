.PHONY: swagger test run dev dev-fast fmt clean-swagger

swagger: clean-swagger
	go generate ./...

test:
	go test ./...

run:
	go run ./cmd/server

dev: swagger test run

dev-fast: test run

fmt:
	gofmt -w .

clean-swagger:
	rm -f docs/docs.go docs/swagger.json docs/swagger.yaml
