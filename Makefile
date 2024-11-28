
test:
	go test -cover ./... -coverprofile coverage.txt
	go test -cover ./... -tags testrunmain
	go tool cover -html coverage.txt
