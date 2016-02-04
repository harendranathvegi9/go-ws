test: lint vet
	go test --race -v .
lint:
	golint .
	test -z "$$(golint .)"
vet:
	go vet .
