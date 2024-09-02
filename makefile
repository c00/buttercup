build: 
	go build -o bin/buttercup cmd/*.go 

install: 
	go install cmd/*.go 

test:
	go test ./...