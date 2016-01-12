test:
	go test -v github.com/marcuswestin/go-ws

test-repeat:
	set -e && while true; do make test; done

test-race:
	go test --race -v github.com/marcuswestin/go-ws
