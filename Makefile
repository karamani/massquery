configure:
	gb vendor restore --all

build:
	gofmt -w src/massquery
	go tool vet src/massquery/*.go
	golint src/massquery
	gb test
	gb build