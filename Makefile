.PHONY: run build test lint clean

run:
	go run main.go

build: 
	go build -o bin/uptimer main.go

test:
	go test ./...

lint:
	go vet ./...

clean:
	rm -f bin/uptimer uptimer.db
	