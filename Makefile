.PHONY: generate test

generate:
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.6.0 -config oapi-codegen.yaml api/openapi.yaml
	go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1 generate

test:
	go test ./...
