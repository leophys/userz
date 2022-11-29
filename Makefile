GO ?= go
COMPOSE ?= docker compose
DBPORT ?= 5432
DLV ?= dlv

.PHONY: test
test:
	$(GO) test -v ./...

.PHONY: test-integration
test-integration:
	$(COMPOSE) -f tests/docker-compose.yaml run --rm tester && \
		$(COMPOSE) -f tests/docker-compose.yaml down

.PHONY: test-clean
test-clean:
	$(COMPOSE) -f tests/docker-compose.yaml down

.PHONY: dbg-integration
dbg-integration:
	ID=$$($(COMPOSE) -f tests/docker-compose.yaml run --rm -d -p 5432:$(DBPORT) db) && \
	GOTAGS="-tag='integration'" $(DLV) test --build-flags="github.com/leophys/userz/tests/store/pg" && \
	docker kill $${ID} && \
	make test-clean
