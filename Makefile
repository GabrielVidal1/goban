BINARY  := kanban-ui
UI_DIR  := ui

.PHONY: build build-ui run run-local dev dev-go dev-ui clean format

format:
	gofmt -w .

build: build-ui
	go mod tidy
	CGO_ENABLED=0 go build -o $(BINARY) .
	@if [ "$$(uname)" = "Darwin" ]; then codesign --force --sign - $(BINARY); fi

build-ui:
	cd $(UI_DIR) && npm install --frozen-lockfile && npm run build

run: build
	sudo -E ./$(BINARY)

run-local: build
	./$(BINARY)

dev:
	@$(MAKE) -j2 dev-go dev-ui

dev-go:
	go run .

dev-ui:
	cd $(UI_DIR) && npm run dev

clean:
	rm -f $(BINARY)
	rm -rf $(UI_DIR)/dist
