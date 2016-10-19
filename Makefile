# Testing
#########

test: lint vet run-tests
ci-test: vet run-tests

run-tests:
	go test --race -v .
lint:
	golint .
	test -z "$$(golint .)"
vet:
	go vet .

# Dependencies
##############

install-golint:
	go get -u github.com/golang/lint/golint