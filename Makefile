cover:
	go test -coverprofile=coverage.out

coverhtml:
	go tool cover -html=coverage.out -o coverage.html

coverfunc:
	go tool cover -func=coverage.out
