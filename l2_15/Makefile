build_tests:
	echo "build tests"
	go build -o example/input example/input.go
	go build -o example/minute example/minute.go
	go build -o example/simple_echo example/simple_echo.go

test: build_tests
	echo "test..."
	go test -v -timeout=30s

	echo "clear..."
	rm -f example/input example/minute example/simple_echo
	rm -f /tmp/minute_output.txt /tmp/test_output.txt

	echo "finish!"