# Defaults — override on the command line if needed:
#   make build WHISPER_BIN=/path/to/whisper-cli MODEL_PATH=/path/to/ggml-small.bin
WHISPER_BIN  ?= $(shell which whisper-cli 2>/dev/null)
MODEL_PATH   ?= $(HOME)/.lay/models/ggml-small.bin
HOMEBREW_LIB ?= /opt/homebrew/lib

APP_RESOURCES := build/bin/lay.app/Contents/Resources

# Dylibs whisper-cli links against via @rpath.
# @loader_path/../lib from Contents/Resources/whisper-cli resolves to Contents/lib/,
# so we copy them there and ad-hoc re-sign so macOS accepts the modified binary.
WHISPER_DYLIBS := \
	libwhisper.1.dylib \
	libggml.0.dylib \
	libggml-cpu.0.dylib \
	libggml-blas.0.dylib \
	libggml-metal.0.dylib \
	libggml-base.0.dylib

.PHONY: build dev clean

build:
	@echo "→ Building lay…"
	wails build
	@if [ -z "$(WHISPER_BIN)" ]; then \
		echo "ERROR: whisper-cli not found. Run: brew install whisper-cpp"; exit 1; \
	fi
	@if [ ! -f "$(MODEL_PATH)" ]; then \
		echo "ERROR: model not found at $(MODEL_PATH)."; \
		echo "  Run: mkdir -p ~/.lay/models && curl -L -o ~/.lay/models/ggml-small.bin \\"; \
		echo "    https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-small.bin"; \
		exit 1; \
	fi
	@echo "→ Bundling whisper-cli and dylibs…"
	mkdir -p "$(APP_RESOURCES)/models"
	mkdir -p "$(APP_RESOURCES)/../lib"
	cp "$(WHISPER_BIN)" "$(APP_RESOURCES)/whisper-cli"
	cp "$(MODEL_PATH)"  "$(APP_RESOURCES)/models/ggml-small.bin"
	@for lib in $(WHISPER_DYLIBS); do \
		cp "$(HOMEBREW_LIB)/$$lib" "$(APP_RESOURCES)/../lib/$$lib"; \
	done
	codesign --force --deep --sign - "$(APP_RESOURCES)/whisper-cli"
	@echo "→ Done. App: $(APP_RESOURCES)/../.."

dev:
	wails dev

clean:
	rm -rf build/bin
