# Defaults — override on the command line if needed:
#   make build WHISPER_BIN=/path/to/whisper-cli MODEL_PATH=/path/to/ggml-small.bin
WHISPER_BIN ?= $(shell which whisper-cli 2>/dev/null)
MODEL_PATH   ?= $(HOME)/.lay/models/ggml-small.bin

APP_RESOURCES := build/bin/lay.app/Contents/Resources

.PHONY: build dev clean

build:
	@echo "→ Building lay…"
	wails build
	@echo "→ Bundling whisper-cli…"
	@if [ -z "$(WHISPER_BIN)" ]; then \
		echo "ERROR: whisper-cli not found. Run: brew install whisper-cpp"; exit 1; \
	fi
	@if [ ! -f "$(MODEL_PATH)" ]; then \
		echo "ERROR: model not found at $(MODEL_PATH)."; \
		echo "  Run: mkdir -p ~/.lay/models && curl -L -o ~/.lay/models/ggml-small.bin \\"; \
		echo "    https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-small.bin"; \
		exit 1; \
	fi
	mkdir -p "$(APP_RESOURCES)/models"
	cp "$(WHISPER_BIN)" "$(APP_RESOURCES)/whisper-cli"
	cp "$(MODEL_PATH)"  "$(APP_RESOURCES)/models/ggml-small.bin"
	@echo "→ Done. App: $(APP_RESOURCES)/../.."

dev:
	wails dev

clean:
	rm -rf build/bin
