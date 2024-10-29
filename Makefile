DOCKER_YAML=-f docker-compose.yml
DOCKER=COMPOSE_PROJECT_NAME=clean-serverless-book-sample docker compose $(DOCKER_YAML)

build:
	$(DOCKER) build ${ARGS}

go-lint:
	$(DOCKER) run go-test ./scripts/go-lint.sh

go-test:
	$(DOCKER) run go-test ./scripts/go-test.sh '${PACKAGE}' '${ARGS}'

go-build:
	$(DOCKER) run go-test ./scripts/build-handlers.sh

go-get:
	$(DOCKER) run go-test go get ${ARGS}
