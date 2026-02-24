# Defaults — override on the command line if needed:
#   make build WHISPER_BIN=/path/to/whisper-cli MODEL_PATH=/path/to/ggml-small.bin
WHISPER_BIN ?= $(shell which whisper-cli 2>/dev/null)
MODEL_PATH  ?= $(HOME)/.lay/models/ggml-small.bin

APP_RESOURCES := build/bin/lay.app/Contents/Resources

# Dylibs whisper-cli links against via @rpath.
# We resolve the actual dylib directory from the binary's own LC_RPATH entry
# (substituting @loader_path with the binary's real directory after symlink
# resolution), then copy them to Contents/lib/ where @loader_path/../lib
# will resolve when whisper-cli runs from Contents/Resources/.
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
	@REAL=$$(readlink -f "$(WHISPER_BIN)"); \
	RPATH=$$(otool -l "$$REAL" | awk '/LC_RPATH/{f=1} f && /path/{print $$2; exit}'); \
	DYLIB_DIR="$${RPATH/@loader_path/$$(dirname $$REAL)}"; \
	for lib in $(WHISPER_DYLIBS); do \
		cp "$$DYLIB_DIR/$$lib" "$(APP_RESOURCES)/../lib/"; \
	done
	codesign --force --deep --sign - "$(APP_RESOURCES)/whisper-cli"
	@echo "→ Done. App: $(APP_RESOURCES)/../.."

dev:
	wails dev

clean:
	rm -rf build/bin
