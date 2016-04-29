configure:
	gb vendor restore --all

build:
	gofmt -w src/massquery
	go tool vet src/massquery/*.go
	gb test
	gb build