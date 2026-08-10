.PHONY: help fmt test build install example-project example-module example-ctrl example-service example-dao clean

APP := gap
BIN := bin/$(APP)
EXAMPLE_ROOT := /tmp/go-agent-platform-example

help:
	@echo "make fmt                         格式化 CLI 工程里的 Go 代码"
	@echo "make test                        运行 CLI 自身单元测试"
	@echo "make build                       编译 gap 脚手架二进制到 bin/gap"
	@echo "make install                     安装 gap 到当前 Go bin 目录"
	@echo "make example-project             在 /tmp 生成一个完整平台工程示例"
	@echo "make example-module NAME=orders  在示例工程根目录生成完整业务模块"
	@echo "make example-ctrl NAME=orders    在示例工程根目录生成 api 和 controller"
	@echo "make example-service NAME=orders 从根级 logic 生成 service 接口"
	@echo "make example-dao NAME=orders     在根级 internal/dao 生成 dao 文件"
	@echo "make clean                       删除 CLI 编译产物"

fmt:
	gofmt -w $$(find . -name '*.go' -print)

test:
	GOCACHE=/tmp/go-agent-platform-cli-cache go test ./...

build:
	GOCACHE=/tmp/go-agent-platform-cli-cache go build -o $(BIN) ./cmd/gap

install:
	GOCACHE=/tmp/go-agent-platform-cli-cache go install ./cmd/gap

example-project:
	go run ./cmd/gap project -name go-agent-platform -module example.com/go-agent-platform -dir $(EXAMPLE_ROOT) -force

example-module:
	go run ./cmd/gap module -name $${NAME:-orders} -root $${ROOT:-$(EXAMPLE_ROOT)} -force

example-ctrl:
	go run ./cmd/gap ctrl -name $${NAME:-orders} -root $${ROOT:-$(EXAMPLE_ROOT)} -force

example-service:
	go run ./cmd/gap service -name $${NAME:-orders} -root $${ROOT:-$(EXAMPLE_ROOT)} -force

example-dao:
	go run ./cmd/gap dao -name $${NAME:-orders} -root $${ROOT:-$(EXAMPLE_ROOT)} -force

clean:
	rm -rf bin
