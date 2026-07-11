SHELL := /bin/bash
.DEFAULT_GOAL := help

ROOT_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))

.PHONY: help local-mq-up local-mq-down local-db-up local-db-down local-sql-load local-im-up local-im-down local-api-chat-run local-api-chat-up local-api-chat-down local-api-chat-admin-run local-api-chat-admin-up local-api-chat-admin-down local-admin-chat-dev local-admin-chat-up local-admin-chat-down local-admin-chat-build local-compose-check swagger-check deploy-mq-up deploy-mq-down deploy-db-up deploy-db-down deploy-sql-load deploy-im-up deploy-im-down deploy-api-chat-up deploy-api-chat-down deploy-api-chat-admin-up deploy-api-chat-admin-down deploy-admin-chat-up deploy-admin-chat-down deploy-all deploy-compose-check im-up im-down mq-up api-chat-run api-chat-admin-run chat-dev chat-build compose-check scan-clean

help:
	@echo ""
	@echo "pte-live-im — 独立 IM 项目集合"
	@echo ""
	@echo "本地 local-*："
	@echo "  make local-mq-up                 启动 MQ/Pulsar，并创建 pte_live_net"
	@echo "  make local-db-up                 启动 MySQL + Redis，加入 pte_live_net"
	@echo "  make local-sql-load              导入 pte-live-sql/001_im_schema.sql"
	@echo "  make local-im-up                 Docker 启动 api-im（固定 IP）"
	@echo "  make local-api-chat-up           Docker 启动 api-chat（固定 IP）"
	@echo "  make local-api-chat-admin-up     Docker 启动 api-chat-admin（固定 IP）"
	@echo "  make local-admin-chat-up         Docker 启动 admin-chat（固定 IP）"
	@echo "  make local-api-chat-run          本地直接运行 api-chat"
	@echo "  make local-api-chat-admin-run    本地直接运行 api-chat-admin"
	@echo "  make local-admin-chat-dev        启动 admin-chat 前端开发服务"
	@echo ""
	@echo "部署 deploy-*："
	@echo "  make deploy-mq-up                部署 MQ/Pulsar"
	@echo "  make deploy-db-up                部署 MySQL + Redis"
	@echo "  make deploy-sql-load             导入 IM SQL"
	@echo "  make deploy-im-up                部署 api-im"
	@echo "  make deploy-api-chat-up          部署 api-chat"
	@echo "  make deploy-api-chat-admin-up    部署 api-chat-admin"
	@echo "  make deploy-admin-chat-up        部署 admin-chat"
	@echo "  make deploy-all                  按 MQ -> DB -> SQL -> IM 服务顺序部署"
	@echo ""
	@echo "校验："
	@echo "  make local-compose-check         校验全部 compose"
	@echo "  make swagger-check               校验 Swagger 覆盖全部 HTTP 路由"
	@echo "  make scan-clean                  扫描不应提交的缓存文件"
	@echo ""
local-mq-up:
	@$(MAKE) -C "$(ROOT_DIR)/pte-live-mq" local-up

local-mq-down:
	@$(MAKE) -C "$(ROOT_DIR)/pte-live-mq" local-down

local-db-up:
	@$(MAKE) -C "$(ROOT_DIR)/pte-live-db" local-up

local-db-down:
	@$(MAKE) -C "$(ROOT_DIR)/pte-live-db" local-down

local-sql-load:
	@$(MAKE) -C "$(ROOT_DIR)/pte-live-db" local-sql-load

local-im-up:
	@$(MAKE) -C "$(ROOT_DIR)/pte-live-im" local-up

local-im-down:
	@$(MAKE) -C "$(ROOT_DIR)/pte-live-im" local-down

local-api-chat-run:
	@$(MAKE) -C "$(ROOT_DIR)/pte-live-api-chat" run

local-api-chat-up:
	@$(MAKE) -C "$(ROOT_DIR)/pte-live-api-chat" local-up

local-api-chat-down:
	@$(MAKE) -C "$(ROOT_DIR)/pte-live-api-chat" local-down

local-api-chat-admin-run:
	@$(MAKE) -C "$(ROOT_DIR)/pte-live-api-chat-admin" run

local-api-chat-admin-up:
	@$(MAKE) -C "$(ROOT_DIR)/pte-live-api-chat-admin" local-up

local-api-chat-admin-down:
	@$(MAKE) -C "$(ROOT_DIR)/pte-live-api-chat-admin" local-down

local-admin-chat-dev:
	cd "$(ROOT_DIR)/pte-live-chat" && pnpm -F @pte/admin-chat run dev

local-admin-chat-up:
	@$(MAKE) -C "$(ROOT_DIR)/pte-live-chat/admin-chat" local-up

local-admin-chat-down:
	@$(MAKE) -C "$(ROOT_DIR)/pte-live-chat/admin-chat" local-down

local-admin-chat-build:
	cd "$(ROOT_DIR)/pte-live-chat" && pnpm -F @pte/admin-chat run build

local-compose-check:
	@docker compose -f "$(ROOT_DIR)/pte-live-mq/docker-compose.yaml" config --quiet
	@docker compose -f "$(ROOT_DIR)/pte-live-db/docker-compose.yaml" config --quiet
	@docker compose -f "$(ROOT_DIR)/pte-live-im/docker-compose.yaml" config --quiet
	@docker compose -f "$(ROOT_DIR)/pte-live-api-chat/docker-compose.yaml" config --quiet
	@docker compose -f "$(ROOT_DIR)/pte-live-api-chat-admin/docker-compose.yaml" config --quiet
	@docker compose -f "$(ROOT_DIR)/pte-live-chat/admin-chat/docker-compose.yaml" config --quiet

swagger-check:
	@ruby "$(ROOT_DIR)/scripts/swagger_check.rb"

# deploy-* 与 local-* 同构，方便服务器脚本统一调用。
deploy-mq-up:
	@$(MAKE) -C "$(ROOT_DIR)/pte-live-mq" deploy-up

deploy-mq-down:
	@$(MAKE) -C "$(ROOT_DIR)/pte-live-mq" deploy-down

deploy-db-up:
	@$(MAKE) -C "$(ROOT_DIR)/pte-live-db" deploy-up

deploy-db-down:
	@$(MAKE) -C "$(ROOT_DIR)/pte-live-db" deploy-down

deploy-sql-load:
	@$(MAKE) -C "$(ROOT_DIR)/pte-live-db" deploy-sql-load

deploy-im-up:
	@$(MAKE) -C "$(ROOT_DIR)/pte-live-im" deploy-up

deploy-im-down:
	@$(MAKE) -C "$(ROOT_DIR)/pte-live-im" deploy-down

deploy-api-chat-up:
	@$(MAKE) -C "$(ROOT_DIR)/pte-live-api-chat" deploy-up

deploy-api-chat-down:
	@$(MAKE) -C "$(ROOT_DIR)/pte-live-api-chat" deploy-down

deploy-api-chat-admin-up:
	@$(MAKE) -C "$(ROOT_DIR)/pte-live-api-chat-admin" deploy-up

deploy-api-chat-admin-down:
	@$(MAKE) -C "$(ROOT_DIR)/pte-live-api-chat-admin" deploy-down

deploy-admin-chat-up:
	@$(MAKE) -C "$(ROOT_DIR)/pte-live-chat/admin-chat" deploy-up

deploy-admin-chat-down:
	@$(MAKE) -C "$(ROOT_DIR)/pte-live-chat/admin-chat" deploy-down

deploy-all: deploy-mq-up deploy-db-up deploy-sql-load deploy-im-up deploy-api-chat-up deploy-api-chat-admin-up deploy-admin-chat-up

deploy-compose-check: local-compose-check

# 旧命令兼容入口，逐步迁移到 local-* / deploy-*。
im-up: local-im-up
im-down: local-im-down
mq-up: local-mq-up
api-chat-run: local-api-chat-run
api-chat-admin-run: local-api-chat-admin-run
chat-dev: local-admin-chat-dev
chat-build: local-admin-chat-build
api-chat-up: local-api-chat-up
api-chat-admin-up: local-api-chat-admin-up
chat-up: local-admin-chat-up
compose-check: local-compose-check

scan-clean:
	@find "$(ROOT_DIR)" \( -name .DS_Store -o -name .gomodcache -o -name .gocache -o -name node_modules -o -name .turbo \) -print
