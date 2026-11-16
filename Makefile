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
cover: ## 生成测试覆盖率报告（根模块单测 + tests/ 集成测试，依赖 Redis）
	go test -race -coverpkg=github.com/jefferyjob/go-redislock -coverprofile=coverage-unit.txt -covermode=atomic ./...
	cd tests && go test -race -coverprofile=../coverage-integration.txt -covermode=atomic -coverpkg=github.com/jefferyjob/go-redislock ./...
	@echo "mode: atomic" > coverage.txt
	@awk '!/^mode:/ { if (!($$1 in stmts)) { order[++n] = $$1; stmts[$$1] = $$2 } cnt[$$1] += $$3 } END { for (i = 1; i <= n; i++) print order[i], stmts[order[i]], cnt[order[i]] }' coverage-unit.txt coverage-integration.txt >> coverage.txt
	@go tool cover -func=coverage.txt
	@rm -f coverage-unit.txt coverage-integration.txt

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