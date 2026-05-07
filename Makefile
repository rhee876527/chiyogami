.PHONY: build-standalone build-embedded check-versions clean

PROJECT := $(CURDIR)
BINARY_STANDALONE := chiyogami-standalone
BINARY_EMBEDDED := chiyogami
BUILD_DIR ?= $(shell mktemp -d)
BUILD_DIR := $(BUILD_DIR)
export BUILD_DIR

build-standalone: check-versions
	# Generate Tailwind CSS output
	cd $(PROJECT)/frameworks && npm ci --silent
	printf '@tailwind base;\n@tailwind components;\n@tailwind utilities;' > $(PROJECT)/public/tailwind.css
	cd $(PROJECT)/frameworks && npx tailwindcss -i $(PROJECT)/public/tailwind.css -o $(PROJECT)/public/output.css --minify

	# Build
	cd $(PROJECT) && go mod tidy && CGO_ENABLED=1 CC=$(CC) GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(BINARY_STANDALONE) . 2>&1
	chmod +x $(PROJECT)/$(BINARY_STANDALONE)

	@echo "Build complete: $(PROJECT)/$(BINARY_STANDALONE)"


build-embedded: check-versions
	@echo "Building in $(BUILD_DIR)"

	# Generate Tailwind CSS output
	cd $(PROJECT)/frameworks && npm ci --silent
	printf '@tailwind base;\n@tailwind components;\n@tailwind utilities;' > $(PROJECT)/public/tailwind.css
	cd $(PROJECT)/frameworks && npx tailwindcss -i $(PROJECT)/public/tailwind.css -o $(PROJECT)/public/output.css --minify

	# Copy source files to build directory
	mkdir -p $(BUILD_DIR)/public $(BUILD_DIR)/core $(BUILD_DIR)/db $(BUILD_DIR)/models
	cp $(PROJECT)/go.mod $(PROJECT)/go.sum $(PROJECT)/LICENSE $(BUILD_DIR)/
	cp -r $(PROJECT)/.git $(PROJECT)/public $(PROJECT)/db $(PROJECT)/models $(PROJECT)/core $(BUILD_DIR)/

	# Create embedded main.go from template
	cp $(PROJECT)/main.go.embedded $(BUILD_DIR)/main.go

	# Create embedded core/embed.go template file (for template parsing)
	printf 'package core\n\nimport "embed"\n\n//go:embed public\nvar Assets embed.FS\n' > $(BUILD_DIR)/core/embed.go

	# Copy public to core for embed access
	cp -r $(BUILD_DIR)/public $(BUILD_DIR)/core/public

	# Fix template parsing in logic.go to use ParseFS with Lookup
	perl -pi -e 's|template\.New\("paste"\)\.ParseFiles\("./public/tmpl\.html"\)|template.ParseFS(Assets, "public/tmpl.html")|g' $(BUILD_DIR)/core/logic.go
	perl -pi -e 's|tmpl\.Execute\(w,|tmpl.Lookup("paste").Execute(w,|g' $(BUILD_DIR)/core/logic.go

	# Build
	cd $(BUILD_DIR) && go mod tidy && CGO_ENABLED=1 CC=$(CC) GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(BINARY_EMBEDDED) . 2>&1

	# Copy binary to project root
	cp $(BUILD_DIR)/$(BINARY_EMBEDDED) $(PROJECT)/$(BINARY_EMBEDDED)
	chmod +x $(PROJECT)/$(BINARY_EMBEDDED)

	# Cleanup build dir on success
	rm -rf $(BUILD_DIR)

	@echo "Build complete: $(PROJECT)/$(BINARY_EMBEDDED)"

check-versions:
	@echo "Checking installed go and node versions..."
	@GO_VERSION=$$(go version 2>/dev/null | grep -oE 'go[0-9]+\.[0-9]+' | sed 's/go//'); \
	if [ "$$GO_VERSION" != "1.26" ]; then \
		echo "WARNING: Go version 1.26 required, found $$GO_VERSION"; \
		exit 1; \
	fi
	@NODE_VERSION=$$(node --version 2>/dev/null | sed 's/v//'); \
	NODE_MAJOR=$$(echo "$$NODE_VERSION" | cut -d. -f1); \
	if [ "$$NODE_MAJOR" != "22" ]; then \
		echo "WARNING: Node version 22 required, found $$NODE_VERSION"; \
		exit 1; \
	fi
	@echo "Versions supported. OK"

clean:
	rm -f $(PROJECT)/$(BINARY_STANDALONE)
	rm -f $(PROJECT)/$(BINARY_EMBEDDED)
	rm -f $(PROJECT)/public/output.css