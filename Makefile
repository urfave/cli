test:
	go test a_test.go $(find . -name '*.go' | egrep -v '_test.go$')

