.PHONY: docker
docker:
	@rm go-postery || true
	@GOOS=linux GOARCH=arm64 go build -o go-postery .
	@docker rmi -f yzletter/go-postery:v0.0.1
	@docker build -t yzletter/go-postery:v0.0.1 .