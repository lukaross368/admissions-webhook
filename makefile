PROJECT_NAME := admissions-webhook
BIN_DIR := bin
CMD_DIR := cmd/webhook
BINARY := $(BIN_DIR)/webhook
CERT_DIR := certs

.PHONY: all
all: build

.PHONY: build
build:
	@echo "🔨 Building $(PROJECT_NAME)..."
	@mkdir -p $(BIN_DIR)
	@go build -ldflags="-s -w" -o $(BINARY) ./$(CMD_DIR)
	@echo "✅ Binary built at $(BINARY)"

.PHONY: build-linux
build-linux:
	@echo "🌍 Building Linux binary..."
	@GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o $(BIN_DIR)/webhook-linux ./$(CMD_DIR)
	@echo "✅ Linux binary built at $(BIN_DIR)/webhook-linux"

.PHONY: certs
certs:
	@echo "🔐 Generating self-signed PEM certs..."
	@mkdir -p $(CERT_DIR)
	@openssl req -x509 -newkey rsa:4096 -sha256 -days 365 -nodes \
		-keyout $(CERT_DIR)/tls.key.pem \
		-out $(CERT_DIR)/tls.crt.pem \
		-subj "/CN=localhost" > /dev/null 2>&1
	@echo "✅ Certs created in $(CERT_DIR)/"

.PHONY: run
run: build certs
	@echo "🚀 Running webhook server with local PEM certs..."
	@go run ./$(CMD_DIR) \
	  --tls-cert-file=$(CERT_DIR)/tls.crt.pem \
	  --tls-private-key-file=$(CERT_DIR)/tls.key.pem \
	  --port=8443

.PHONY: test
test:
	@echo "🧪 Running all unit tests for $(PROJECT_NAME)..."
	@go test ./... -v -cover
	@echo "✅ All tests passed!"

.PHONY: clean
clean:
	@echo "🧹 Cleaning up binaries and certs..."
	@rm -rf $(BIN_DIR) $(CERT_DIR)
	@echo "✅ Cleanup complete"