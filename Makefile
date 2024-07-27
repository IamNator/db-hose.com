run:
	go run main.go 

build: 
	docker build -t db-hose .

fmt:
	go fmt ./...