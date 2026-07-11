# pte-live-api-chat — 聊天与场景消息 API

SHELL := /bin/bash
.DEFAULT_GOAL := help

CHAT_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
CHAT_API_PORT ?= 11504
BIN := $(CHAT_DIR)/pte_live_api_chat
COMPOSE := docker compose -p pte_live_api_chat -f $(CHAT_DIR)/docker-compose.yaml

export CHAT_API_PORT
export DOCKER_BUILDKIT := 1
IMAGE_TAG ?= v1
export IMAGE_TAG

.PHONY: help local-tidy local-build local-run local-test local-up local-down local-restart local-logs local-ps deploy-build deploy-up deploy-down deploy-restart deploy-logs deploy-ps tidy build run test up down restart deploy logs ps docker-build

help: ## 显示可用命令
	@echo ""
	@echo "pte_live_api_chat — 聊天与场景消息 API（端口 $(CHAT_API_PORT)）"
	@echo ""
	@echo "本地 local-*：local-run/local-test/local-up/local-down"
	@echo "部署 deploy-*：deploy-build/deploy-up/deploy-down"
	@echo "旧命令兼容：run/build/test/deploy/up/down/logs/ps"
	@echo ""

local-tidy: ## go mod tidy
	cd $(CHAT_DIR) && go mod tidy

local-build: ## 本地编译
	cd $(CHAT_DIR) && go build -o $(BIN) ./cmd/server

local-run: local-build ## 本地直接运行
	cd $(CHAT_DIR) && ./pte_live_api_chat -c conf/app.yaml.example

local-test: ## 运行测试
	cd $(CHAT_DIR) && go test ./...

local-up: deploy-up ## Docker 启动，固定 IP

local-down: deploy-down ## Docker 停止

local-restart: deploy-restart ## Docker 重启

local-logs: deploy-logs ## Docker 日志

local-ps: deploy-ps ## Docker 状态

deploy-build: ## 构建 Docker 镜像
	$(COMPOSE) build

deploy-up: deploy-build ## Docker 部署启动
	$(COMPOSE) up -d


deploy-down: ## Docker 停止
	$(COMPOSE) down

deploy-restart: ## Docker 重启
	$(COMPOSE) restart

deploy-logs: ## Docker 日志
	$(COMPOSE) logs -f api-chat

deploy-ps: ## Docker 状态
	$(COMPOSE) ps -a

# 旧命令兼容入口
tidy: local-tidy
build: local-build
run: local-run
test: local-test
docker-build: deploy-build
deploy: deploy-up
up: deploy-up
down: deploy-down
restart: deploy-restart
logs: deploy-logs
ps: deploy-ps
