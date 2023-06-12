build:
	@mkdir -p ./bin
	@go build -o ./bin/awsenv main.go access.go

clean:
	@rm -rf ./bin