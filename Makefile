VERSION ?= $(shell git describe --tags --always)
GOBIN ?= $(shell go env GOPATH)/bin

run:
	go run \
		-ldflags="-X main.BuildVersion=$(VERSION)" \
		./cmd/java-tuner \
		--verbose \
		--dry-run

bin/java-tuner: build

build:
	CGO_ENABLED=0 go build \
		-ldflags="-X main.BuildVersion=$(VERSION)" \
		-o bin/java-tuner ./cmd/java-tuner

test: bin/java-tuner docker-run
	@echo "Running tests..."
	go test -v ./...

$(GOBIN)/goimports:
	@go install golang.org/x/tools/cmd/goimports@v0.32.0

$(GOBIN)/gocyclo:
	@go install github.com/fzipp/gocyclo/cmd/gocyclo@v0.6.0

$(GOBIN)/golangci-lint:
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.3.0

$(GOBIN)/gocritic:
	@go install github.com/go-critic/go-critic/cmd/gocritic@v0.13.0

install-linters: $(GOBIN)/goimports $(GOBIN)/gocyclo $(GOBIN)/golangci-lint $(GOBIN)/gocritic
	@echo "Linters installed successfully."

lint: install-linters
	@pre-commit run -a

docker-build:
	docker build -t java-tuner:$(VERSION) .

docker-run: docker-build
	docker run -ti --rm --cpu-quota 1000 java-tuner:$(VERSION) --dry-run
	docker run -ti --rm --cpu-quota 1000000 java-tuner:$(VERSION) --verbose
	docker run -ti --rm -e JAVA_TUNER_CPU=2 java-tuner:$(VERSION) --verbose
	docker run -ti --rm -m 128m java-tuner:$(VERSION) --verbose --dry-run
	docker run -ti --rm -m 32m java-tuner:$(VERSION) --verbose --log-format json
	docker run -ti --rm -m 128m -e JAVA_TUNER_MEM_PERCENTAGE=50 java-tuner:$(VERSION) --verbose
	docker run -ti --rm -m 32m java-tuner:$(VERSION) --verbose -- -jar ./app.jar

	docker run -ti --rm -m 64m \
		-e JAVA_TUNER_CPU_COUNT=2 \
		-e JAVA_TUNER_OPTS="-XX:MaxRAMFraction=1" \
		-e JAVA_TUNER_DRY_RUN=false \
		-e JAVA_TUNER_LOG_FORMAT="plain" \
		java-tuner:$(VERSION) --verbose

clean:
	@rm -rfv bin
	@find example -name '*.Dockerfile' -delete
	@find tests -name '*.Dockerfile' -delete
