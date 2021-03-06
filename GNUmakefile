PKG_NAME	=		api-gateway
IMG_NAME 	= 	api-gateway
FILES			?=	$$(find . -name '*.go' )
CERTS 		= ./certs

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

cert:
	# @openssl genrsa -out ${CERTS}/private.key 4096
	# @openssl req -new -x509 -sha256 -days 1825 -addext "subjectAltName = DNS:localhost" \
	# -key ${CERTS}/private.key -out ${CERTS}/public.crt
	@./scripts/certgen.sh

tidy:
	@ go mod tidy --compat=1.18
	@ echo "Done!"

upd:
	@ ./scripts/update.sh
	@ echo "Done!"

apigen:
	@./scripts/apigen.sh

protogen:
	@./scripts/protogen.sh

run:
	@go run -v ./cmd/proto/main.go

runrest:
	@go run -v ./cmd/rest/main.go

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
	@./scripts/build.sh	

.PHONY: test lint vet fmt protogen cert
