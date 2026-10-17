# Plugin ID = file name of the .ndp package (without extension).
# Keep it stable: Navidrome stores config, user access and KV data per plugin ID.
PLUGIN_ID := album-rating

.PHONY: all test build package clean tidy docker

all: package

tidy:
	go mod tidy

test:
	go test ./...

build:
	GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared -o plugin.wasm .

package: build
	zip -j $(PLUGIN_ID).ndp manifest.json plugin.wasm

# Build inside a container – no local Go installation required.
docker:
	docker run --rm -v "$(CURDIR)":/src -w /src golang:1.26 sh -c "\
	  apt-get update -qq && apt-get install -y -qq zip >/dev/null && \
	  go mod tidy && go test ./... && \
	  GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared -o plugin.wasm . && \
	  zip -j $(PLUGIN_ID).ndp manifest.json plugin.wasm"

clean:
	rm -f plugin.wasm *.ndp
	rm -rf dist
