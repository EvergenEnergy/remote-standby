help:
	@grep -E '^[\.a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

test.unit: test.unit.run ## run only unit tests

test.integration: test.integration.setup ## run only integration tests
test.integration: test.integration.run
test.integration: test.integration.teardown

test.all: test.integration.setup ## run all tests
test.all: test.all.run
test.all: test.integration.teardown

test.integration.setup:
	docker compose -f docker-compose.yml up -d

	echo "dependencies started"

test.unit.run:
	gotestsum --junitfile test-reports/junit.xml -- -timeout 1m -count=1 -coverprofile=cp.out -race -short -v ./...

test.integration.run:
	gotestsum --junitfile test-reports/junit.xml -- -timeout 2m -count=1 -coverprofile=cp.out -failfast -race -run Integration -v ./...

test.all.run:
	gotestsum --junitfile test-reports/junit.xml -- -timeout 3m -count=1 -coverprofile=cp.out -race -v ./...

test.integration.teardown:
	docker compose -f docker-compose.yml down -v
	docker compose -f docker-compose.yml rm -s -f -v

run.local: ## run the app locally
	set -o allexport; source .env.local; set +o allexport && go run .
