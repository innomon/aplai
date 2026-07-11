.PHONY: all build dylib clean test

all: build

dylib: libapple_ai_bridge.dylib

libapple_ai_bridge.dylib: apple_ai_bridge.swift
	swiftc -emit-library -o libapple_ai_bridge.dylib apple_ai_bridge.swift -sdk $$(xcrun --show-sdk-path) -Xlinker -install_name -Xlinker @rpath/libapple_ai_bridge.dylib

build: dylib
	CGO_LDFLAGS_ALLOW="-Wl,-rpath,@loader_path/.*" go build -o aplai cmd/aplai/main.go

test: dylib
	CGO_LDFLAGS_ALLOW="-Wl,-rpath,@loader_path/.*" go test -v ./...

clean:
	rm -f libapple_ai_bridge.dylib aplai
