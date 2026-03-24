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
	@scp ./app/linux/code_service myserver1:~/app/code_service
	@scp ./app/linux/auth_service myserver1:~/app/auth_service
	@scp ./app/linux/lottery_service myserver1:~/app/lottery_service
	@scp ./app/linux/agent_service myserver5:~/app/agent_service
	@scp ./app/linux/user_service myserver1:~/app/user_service
	@scp ./app/linux/session_service myserver1:~/app/session_service
	@scp ./app/linux/outbox_service myserver5:~/app/outbox_service
