PKG_NAME	=		nws-sdk-go
IMG_NAME 	= 	nws-sdk-api
FILES			?=	$$(find . -name '*.go' )

default: fmt 

fmt:
	@gofmt -w $(FILES)

vet:
	@echo "Running go vet..."
	@go vet ./...; if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi

tidy:
	@ go mod tidy --compat=1.18
	@ echo "Done!"

upd:
	@ cd internal/cloudcontrol/ && go get -u && go mod tidy --compat=1.18
	@ echo "Done!"

apigen:
	@./scripts/apigen.sh

run:
	@go run -v ./cmd/main.go

test: 
	@ go test -v ./...
	@ echo "Done!"

lint: golangci-lint

golangci-lint:
	@echo "==> Checking source code with golangci-lint..."
	@docker run --rm -v $(PWD):/app -w /app golangci/golangci-lint golangci-lint run --fix ./...

semgrep:
	@echo "==> Running Semgrep static analysis..."
	@docker run --rm -v $(PWD):/src returntocorp/semgrep --config=auto --verbose

image:
	@docker build -f Dockerfile.dev -t tampler/$(IMG_NAME) .

.PHONY: test lint vet fmt
