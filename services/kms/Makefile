mock:
	rm -rf ./test/mocks/* && docker run --rm -i -v ${PWD}:/src -v ${PWD}/.netrc:/root/.netrc -w /src vektra/mockery:3.7

lint:
	docker run --rm -i -w /src -v ${PWD}:/src -v ${PWD}/.netrc:/root/.netrc golangci/golangci-lint:v2.11.4-alpine golangci-lint run

lint-fix:
	docker run --rm -i -w /src -v ${PWD}:/src -v ${PWD}/.netrc:/root/.netrc golangci/golangci-lint:v2.11.4-alpine golangci-lint run --fix

run-tests:
	docker run --rm -i -v ${PWD}:/src -w /src golang:1.26.4 go test ./...

run-bench:
	docker run --rm -i -v ${PWD}:/src -w /src golang:1.26.4 go test -bench=^$ -cpu=4 ./...

db-gen:
	docker run --network none --rm -v ${PWD}:/src -w /src sqlc/sqlc generate
