# A Self-Documenting Makefile: http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY:build
build: ## 编译项目
	go build -v ./...

.PHONY:test
test: ## 运行测试
	go test -v ./...

.PHONY:lint
lint: ## 执行代码静态分析
	golangci-lint run

.PHONY:bench
bench: ## 运行基准测试
	go test -benchmem -bench .

.PHONY:doc
doc: ## 启动文档服务器
	godoc -http=:6060 -play -index

.PHONY:cover
cover: ## 生成测试覆盖率报告
	#go tool cover -func=coverage.out
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...

.PHONY:run-redis
run-redis: ## Docker启动redis服务
	docker run -itd -p 63790:6379 --name example_redislock redis:5.0.3-alpine

.PHONY:mocks
mocks: ## 基于Interface生成Mock代码
	mockgen -source=lock.go -destination=mocks/lock.go -package=mocks

.PHONY:help
.DEFAULT_GOAL:=help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'