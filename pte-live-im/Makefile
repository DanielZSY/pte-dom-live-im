# pte_live_im API service
# MQ/DB are split into ../pte-live-mq and ../pte-live-db.

SHELL := /bin/bash
.DEFAULT_GOAL := help

IM_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
ROOT_DIR := $(abspath $(IM_DIR)/..)

PROJECT_NAME ?= pte_live_api_im

WS_PORT   ?= 11510
HTTP_PORT ?= 11511
RPC_PORT  ?= 11512
IM_DOC_PORT ?= 11552

BUILD_DIR := build
BIN_DIR := $(BUILD_DIR)/dist
BIN     := $(BIN_DIR)/pte_live_im
MAIN    := ./cmd/main
PROJECT_PKGS := ./api/... ./cmd/... ./define/... ./grpcserver/... ./pkg/... ./protobuf/... ./queue/... ./routers/... ./servers/... ./tools/...

COMPOSE_IM := docker compose -p pte_live_api_im -f docker-compose.yaml

export DOCKER_BUILDKIT := 1
IMAGE_TAG ?= v1
export IMAGE_TAG

IM_CONFIG := conf/app.yaml
DEPLOY_DESC := Docker（api-im；MQ/DB 使用同一 pte_live_net）

export WS_PORT HTTP_PORT RPC_PORT IM_CONFIG

.PHONY: help local-deps local-build local-run local-test local-lint local-release local-up local-down local-restart local-logs local-ps local-status local-clean local-clean-dist deploy-build deploy-up deploy-down deploy-restart deploy-logs deploy-ps deploy-status deploy-clean deploy-all deps build run test lint release docker-build deploy up down restart logs ps status clean clean-dist mq-up mq-down

help: ## 显示可用命令
	@echo ""
	@echo "pte_live_api_im"
	@echo ""
	@echo "本地 local-*："
	@echo "  make local-deps        预下载 go mod（使用 Go 默认模块缓存）"
	@echo "  make local-build       编译到 $(BIN)"
	@echo "  make local-run         编译并运行（conf/app.yaml）"
	@echo "  make local-test        运行测试"
	@echo "  make local-up          构建并启动 api-im（要求 pte_live_net 已由 pte-live-mq 创建）"
	@echo ""
	@echo "部署 deploy-*："
	@echo "  make deploy-build      构建 IM 镜像"
	@echo "  make deploy-up         启动 api-im"
	@echo "  make deploy-down       停止 api-im"
	@echo ""
	@echo "兼容旧命令：deps/build/run/test/deploy/up/down/logs/ps/status"
	@echo ""
	@echo "当前 IM_CONFIG=$(IM_CONFIG)（compose 挂载 conf/app.yaml）"
	@echo "服务端口：WS $(WS_PORT)，HTTP $(HTTP_PORT)，gRPC $(RPC_PORT)，Swagger $(IM_DOC_PORT)"
	@echo ""

local-deps: ## 预下载依赖（使用 Go 默认模块缓存）
	go mod download

local-build: ## 编译二进制到 build/dist
	@mkdir -p $(BIN_DIR)
	go build -mod=readonly -o $(BIN) $(MAIN)
	@echo ">> built $(BIN)"

local-run: local-build ## 本地运行（conf/app.yaml；无则 app.yaml.example）
	@if [ -f conf/app.yaml ]; then cfg=conf/app.yaml; else cfg=conf/app.yaml.example; fi && \
	$(BIN) -c "$$cfg"

local-test: ## 运行测试
	go test $(PROJECT_PKGS)

local-lint: ## golangci-lint 代码检查
	golangci-lint run $(PROJECT_PKGS)

local-release: ## goreleaser 本地快照打包到 build/goreleaser
	goreleaser release --snapshot --clean

local-up: deploy-up ## 本地启动 api-im

local-down: deploy-down ## 本地停止 api-im

local-restart: deploy-restart ## 本地重启 api-im

local-logs: deploy-logs ## 查看本地日志

local-ps: deploy-ps ## 查看本地状态

local-status: deploy-status ## 打印本地状态

local-clean: deploy-clean ## 清理 api-im compose

local-clean-dist: clean-dist ## 删除 dist 构建产物

deploy-build: ## 构建 IM 镜像
	$(COMPOSE_IM) build

deploy-up: deploy-build ## 构建并启动 api-im
	@echo ">> 部署: $(DEPLOY_DESC)"
	@$(COMPOSE_IM) up -d
	@$(MAKE) --no-print-directory deploy-status

deploy-all: deploy-up ## deploy-up 别名

deploy-down: ## 停止 api-im
	@$(COMPOSE_IM) down

deploy-restart: ## 重启 api-im
	@$(COMPOSE_IM) restart

deploy-logs: ## 查看 api-im 日志
	$(COMPOSE_IM) logs -f api-im

deploy-ps: ## 查看 api-im 容器状态
	$(COMPOSE_IM) ps -a

deploy-status: ## 打印访问地址与 IM 运行状态
	@echo ""
	@echo "api-im 运行状态:"
	@$(COMPOSE_IM) ps
	@echo ""
	@echo "配置: $(IM_CONFIG)"
	@echo "访问地址:"
	@echo "  WebSocket : ws://127.0.0.1:$(WS_PORT)/ws?sdkAppID=<sdkAppID>&identifier=<identifier>&userSig=<userSig>"
	@echo "  HTTP API  : http://127.0.0.1:$(HTTP_PORT)/api/register"
	@echo "  Live API  : http://127.0.0.1:$(HTTP_PORT)/api/live/*"
	@echo "  gRPC      : 127.0.0.1:$(RPC_PORT)  (ImApi, metadata: appid)"
	@echo "  Swagger   : http://127.0.0.1:$(IM_DOC_PORT)"
	@echo ""

deploy-clean: ## 停止并删除 api-im 容器（不删镜像、网络、MQ、DB）
	$(COMPOSE_IM) down --remove-orphans

clean-dist: ## 删除 dist 构建产物
	rm -rf $(BIN_DIR)

# 旧命令兼容入口，逐步迁移到 local-* / deploy-*。
deps: local-deps
build: local-build
run: local-run
test: local-test
lint: local-lint
release: local-release
docker-build: deploy-build
deploy: deploy-up
up: deploy-up
down: deploy-down
restart: deploy-restart
logs: deploy-logs
ps: deploy-ps
status: deploy-status
clean: deploy-clean
mq-up:
	@$(MAKE) -C "$(ROOT_DIR)/pte-live-mq" local-up
mq-down:
	@$(MAKE) -C "$(ROOT_DIR)/pte-live-mq" local-down
