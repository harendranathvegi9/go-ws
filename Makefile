test: lint vet
	go test -v .
test-race: test
	go test --race -v .
lint:
	golint .
	test -z "$$(golint .)"
vet:
	go vet .
