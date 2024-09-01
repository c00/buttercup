build: 
	go build -o bin/buttercup cmd/*.go 

test:
	go test ./...