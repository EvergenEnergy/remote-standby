help:
	@grep -E '^[\.a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

SHUNIT_VERSION = 2.1.8
SHUNIT_DIR = shunit2

test.unit: test.unit.run ## run only unit tests

test.integration: test.integration.setup ## run only integration tests
test.integration: test.integration.run
test.integration: test.integration.teardown

test.all: test.integration.setup ## run all tests
test.all: test.all.run
test.all: test.integration.teardown

test-prep:
ifeq (,$(wildcard shunit2/shunit2))
ifeq (,$(wildcard v2.1.8.tar.gz))
	@echo "Downloading test framework"
	@wget https://github.com/kward/shunit2/archive/refs/tags/v$(SHUNIT_VERSION).tar.gz
endif
	@rm -r shunit2 || true
	@tar -xf v$(SHUNIT_VERSION).tar.gz && mv shunit2-$(SHUNIT_VERSION) $(SHUNIT_DIR)
endif

test.integration.setup: test-prep
	docker compose -f docker-compose.yaml up -d

	echo "dependencies started"

test.unit.run:
	gotestsum --junitfile test-reports/junit.xml -- -timeout 1m -count=1 -coverprofile=covprofile -race -short -v ./...

test.integration.run:
	gotestsum -- -count=1 -coverprofile=covprofile -failfast -race -run Integration -v ./...
	./tests/integration/end-to-end-tests.sh

test.all.run:
	gotestsum --junitfile test-reports/junit.xml -- -timeout 3m -count=1 -coverprofile=covprofile -race -v ./...

test.integration.teardown:
	docker compose -f docker-compose.yaml down -v
	docker compose -f docker-compose.yaml rm -s -f -v

run.local: ## run the app locally
	set -o allexport; source .env.local; set +o allexport && go run .
