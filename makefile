build:
	@go build -o ./bin/analyzer ./cmd/mylinter

build-plugin:
	@go build -buildmode=plugin -o ./bin/mylinter.so ./plugin

run:
	./bin/analyzer ./example/example.go
