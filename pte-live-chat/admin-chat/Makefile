# pte_live_admin_chat — Vben IM 后台（nginx + /admin/im 反代）

SHELL := /bin/bash
.DEFAULT_GOAL := help

export DOCKER_BUILDKIT := 1
IMAGE_TAG ?= v1
export IMAGE_TAG

ADMIN_CHAT_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
ROOT_DIR := $(abspath $(ADMIN_CHAT_DIR)/../..)

ADMIN_CHAT_PORT ?= 11526
PNPM ?= pnpm
ROOT_VBEN_DIR := $(abspath $(ADMIN_CHAT_DIR)/..)
COMPOSE := docker compose -p pte_live -f docker-compose.yaml

export ADMIN_CHAT_PORT

.PHONY: help install dev pack build preview require-dist local-install local-dev local-pack local-build local-preview local-up local-down local-restart local-logs local-ps deploy-build deploy-up deploy-down deploy-restart deploy-logs deploy-ps deploy docker-build ensure-network ensure-api up down restart logs ps

help: ## 显示可用命令
	@echo ""
	@echo "pte_live_admin_chat — Vben IM 后台"
	@echo ""
	@grep -E '^[a-zA-Z0-9_-]+:.*##' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*## "}; {printf "  %-14s %s\n", $$1, $$2}'
	@echo ""
	@echo "本地开发: http://localhost:$(ADMIN_CHAT_PORT)"
	@echo "生产发布: 本地 make pack → 上传 dist/ 到 $(ADMIN_CHAT_DIR)/dist/ → 服务器 make deploy"
	@echo "Token Key: pteLiveChatAdminToken"
	@echo ""

install: ## 安装 pnpm 依赖（在 monorepo 根目录）
	cd "$(ROOT_VBEN_DIR)" && $(PNPM) install

dev: ## 启动 Vben 开发服务（端口 $(ADMIN_CHAT_PORT)）
	cd "$(ROOT_VBEN_DIR)" && $(PNPM) dev:admin-chat

pack: ## 本地 Vite 打包 → dist/（开发机执行）
	cd "$(ROOT_VBEN_DIR)" && $(PNPM) build:admin-chat

build: pack ## 同 pack

require-dist: ## 检查 dist/ 是否存在
	@test -f "$(ADMIN_CHAT_DIR)/dist/index.html" || { \
		echo "缺少 $(ADMIN_CHAT_DIR)/dist/index.html" >&2; \
		echo "请先在本机 make pack，再将 dist/ 上传到服务器同路径" >&2; \
		exit 1; \
	}

preview: pack ## 本地预览 dist/ 构建产物
	cd "$(ADMIN_CHAT_DIR)" && $(PNPM) exec vite preview --host 0.0.0.0 --port $(ADMIN_CHAT_PORT)

docker-build: require-dist ## 构建 Docker 镜像（仅 COPY dist/）
	$(COMPOSE) build admin-chat

ensure-network: ## 确保 Docker 网络 pte_live_net 已由 pte-live-mq 创建
	@docker network inspect pte_live_net >/dev/null 2>&1 || { \
		echo "缺少 Docker 网络 pte_live_net，请先执行: make -C ../../pte-live-mq local-up"; \
		exit 1; \
	}

ensure-api: ensure-network ## 严格检查 API（失败则 exit 1）
	@curl -sf "http://127.0.0.1:11505/ping" >/dev/null || { \
		echo "api-chat-admin 未就绪，请先执行: make -C ../../pte-live-api-chat-admin run"; \
		exit 1; \
	}

local-install: install ## local-* 别名

local-dev: dev ## local-* 别名

local-pack: pack ## local-* 别名

local-build: pack ## local-* 别名

local-preview: preview ## local-* 别名

deploy-build: docker-build ## 构建 Docker 镜像

deploy-up: ensure-network deploy-build ## 构建镜像并启动（需 dist/ 已存在）
	@$(COMPOSE) up -d
	@echo ">> Vben IM 后台: http://localhost:$(ADMIN_CHAT_PORT)"

deploy-down: ## 停止并移除容器
	@$(COMPOSE) down

deploy-restart: deploy-down deploy-up ## 重新构建并启动

deploy-logs: ## 查看日志
	$(COMPOSE) logs -f admin-chat

deploy-ps: ## 查看容器状态
	$(COMPOSE) ps -a

local-up: deploy-up ## local-* 别名
local-down: deploy-down ## local-* 别名
local-restart: deploy-restart ## local-* 别名
local-logs: deploy-logs ## local-* 别名
local-ps: deploy-ps ## local-* 别名

# 旧命令兼容入口
up: deploy-up
down: deploy-down
restart: deploy-restart
deploy: deploy-up
logs: deploy-logs
ps: deploy-ps
