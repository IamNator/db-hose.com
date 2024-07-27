fmt:
	go fmt ./...

test:
	go test -v ./...

run:
	go run main.go 

build-docker: 
	docker build -t db-hose .

run-docker:
	docker run -p 8080:8080 db-hose