HOSTNAME=registry.terraform.io
NAMESPACE=askrella
NAME=ssh
BINARY=terraform-provider-${NAME}
VERSION=0.1.0
OS_ARCH=darwin_arm64

default: install

build:
	go build -o ${BINARY}

release:
	GOOS=darwin GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_darwin_amd64
	GOOS=darwin GOARCH=arm64 go build -o ./bin/${BINARY}_${VERSION}_darwin_arm64
	GOOS=linux GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_linux_amd64
	GOOS=linux GOARCH=arm64 go build -o ./bin/${BINARY}_${VERSION}_linux_arm64
	GOOS=windows GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_windows_amd64

install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}/${BINARY}

examples: install
	export TF_CLI_CONFIG_FILE=terraform.tfrc \
		&& cd examples \
		&& rm -rf ./.terraform ./.terraform.lock.hcl ./terraform.tfstate ./terraform.tfstate.backup \
		&& terraform init --reconfigure \
		&& terraform plan \
		&& terraform apply --auto-approve #\
		#&& terraform destroy --auto-approve

test: setup-test
	go test -v ./...

.PHONY: setup-test
setup-test:
	mkdir -p mount
	chmod 777 mount

.PHONY: testacc
testacc: setup-test
	TF_ACC=1 go test ./... -v $(TESTARGS)

fmt:
	go fmt ./...

lint:
	golangci-lint run

.PHONY: build release install examples fmt lint test-env-up test-env-down test-with-docker clean

test-env-up:
	docker-compose -f docker-compose.test.yml up -d --build
	@echo "Waiting for SSH server to be ready..."
	@timeout 30s sh -c 'until nc -z localhost 2222; do sleep 1; done' || (make test-env-down && exit 1)

test-env-down:
	docker-compose -f docker-compose.test.yml down -v

test-with-docker: 
	@echo "Starting test environment..."
	@make test-env-up || exit 1
	@echo "Running tests..."
	@go test ./... -v -parallel 1 || (ret=$$?; make test-env-down; exit $$ret)
	@echo "Cleaning up test environment..."
	@make test-env-down

clean:
	rm -rf mount 