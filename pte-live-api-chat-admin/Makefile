# pte-live-api-chat-admin — IM 后台 API

SHELL := /bin/bash
.DEFAULT_GOAL := help

CHAT_ADMIN_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
CHAT_ADMIN_PORT ?= 11505
BIN := $(CHAT_ADMIN_DIR)/pte_live_api_chat_admin
COMPOSE := docker compose -p pte_live_api_chat_admin -f $(CHAT_ADMIN_DIR)/docker-compose.yaml

export CHAT_ADMIN_PORT
export DOCKER_BUILDKIT := 1
IMAGE_TAG ?= v1
export IMAGE_TAG

.PHONY: help local-tidy local-build local-run local-test local-up local-down local-restart local-logs local-ps deploy-build deploy-up deploy-down deploy-restart deploy-logs deploy-ps tidy build run test up down restart deploy logs ps docker-build

help:
	@echo ""
	@echo "pte_live_api_chat_admin — IM 后台 API（端口 $(CHAT_ADMIN_PORT)）"
	@echo ""
	@echo "本地 local-*：local-run/local-test/local-up/local-down"
	@echo "部署 deploy-*：deploy-build/deploy-up/deploy-down"
	@echo "旧命令兼容：run/build/test/deploy/up/down/logs/ps"
	@echo ""

local-tidy:
	cd $(CHAT_ADMIN_DIR) && go mod tidy

local-build:
	cd $(CHAT_ADMIN_DIR) && go build -o $(BIN) ./cmd/server

local-run:
	cd $(CHAT_ADMIN_DIR) && go run ./cmd/server -c conf/app.yaml.example

local-test:
	cd $(CHAT_ADMIN_DIR) && go test ./...

local-up: deploy-up

local-down: deploy-down

local-restart: deploy-restart

local-logs: deploy-logs

local-ps: deploy-ps

deploy-build:
	$(COMPOSE) build

deploy-up: deploy-build
	$(COMPOSE) up -d

deploy-down:
	$(COMPOSE) down

deploy-restart:
	$(COMPOSE) restart

deploy-logs:
	$(COMPOSE) logs -f api-chat-admin

deploy-ps:
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
