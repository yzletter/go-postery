.PHONY: build

build:
	@rm -rf ./app/linux || true
	@go mod tidy
	@GOOS=linux GOARCH=amd64 go build -o ./app/linux/code_service ./microservice-backend/code
	@GOOS=linux GOARCH=amd64 go build -o ./app/linux/auth_service ./microservice-backend/auth
	@GOOS=linux GOARCH=amd64 go build -o ./app/linux/lottery_service ./microservice-backend/lottery
	@GOOS=linux GOARCH=amd64 go build -o ./app/linux/post_service ./microservice-backend/post
	@GOOS=linux GOARCH=amd64 go build -o ./app/linux/agent_service ./microservice-backend/agent
	@GOOS=linux GOARCH=amd64 go build -o ./app/linux/user_service ./microservice-backend/user
	@GOOS=linux GOARCH=amd64 go build -o ./app/linux/session_service ./microservice-backend/session
	@GOOS=linux GOARCH=amd64 go build -o ./app/linux/outbox_service ./microservice-backend/outbox
	@GOOS=linux GOARCH=amd64 go build -o ./app/linux/bff_service ./microservice-backend/bff
