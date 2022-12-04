GO ?= go
DOCKER ?= docker
COMPOSE ?= docker compose
IMAGE ?= leophys/userz
DBPORT ?= 5432
DLV ?= dlv
OUTDIR ?= bin/
VERSION ?= $(shell git rev-list -1 HEAD)

./bin:
	mkdir -p bin

./bin/userz: ./bin
	$(GO) build $(BUILD_OPTS) -o $(OUTDIR) ./cmd/userz/...

./bin/pollednotifier.so: ./bin
	$(GO) build $(BUILD_OPTS) -buildmode=plugin -o $(OUTDIR) ./internal/pollednotifier/...

.PHONY: clean
clean: ./bin
	rm -f bin/userz

.PHONY: build
build: clean
	make ./bin/userz
	make ./bin/pollednotifier.so

.PHONY: prod
prod: clean
	make BUILD_OPTS="-ldflags '-X=main.commit=$(VERSION) -w -s'" build

.PHONY: build-image
build-image:
	$(DOCKER) build \
		--build-arg=BUILD_OPTS="$(BUILD_OPTS)" \
		-f cmd/userz/Dockerfile \
		-t $(IMAGE) .

.PHONY: build-image-prod
build-image-prod:
	make BUILD_OPTS="-ldflags '-X=main.commit=$(VERSION) -w -s'" build-image

.PHONY: run
run:
	$(COMPOSE) build
	$(COMPOSE) up -d
	$(COMPOSE) logs -f userz

.PHONY: data
data:
	./tests/init_db.sh

.PHONY:
stop:
	$(COMPOSE) stop
	$(COMPOSE) down --volumes

.PHONY: test
test:
	$(GO) test -v -count=1 ./...

.PHONY: test-integration
test-integration:
	$(COMPOSE) -f tests/docker-compose.yaml run --rm tester && \
		$(COMPOSE) -f tests/docker-compose.yaml down --volumes

.PHONY: test-clean
test-clean:
	$(COMPOSE) -f tests/docker-compose.yaml down --volumes

.PHONY: dbg-integration
dbg-integration:
	ID=$$($(COMPOSE) -f tests/docker-compose.yaml run --rm -d -p 5432:$(DBPORT) db) && \
	   POSTGRES_URL="postgres://userz:passw0rd@localhost:5432/userz?sslmode=disable" \
	   GOFLAGS="-tags='integration'" \
	   $(DLV) test --build-flags="github.com/leophys/userz/tests/store/pg" && \
	docker kill $${ID} && \
	make test-clean
