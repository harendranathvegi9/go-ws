# Testing
#########

test-ci: test
test: lint vet run-tests

run-tests:
	go test -v .
	go test --race -v .
lint:
	golint -set_exit_status .
vet:
	go vet .

# Dependencies
##############

install-golint:
	go get -u github.com/golang/lint/golint