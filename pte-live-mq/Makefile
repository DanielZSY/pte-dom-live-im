SHELL := /bin/bash
.DEFAULT_GOAL := help

MQ_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
COMPOSE_MQ := docker compose -p pte_live_mq -f $(MQ_DIR)/docker-compose.yaml

.PHONY: help local-up local-down local-restart local-logs local-ps deploy-up deploy-down deploy-restart deploy-logs deploy-ps

help:
	@echo ""
	@echo "pte-live-mq"
	@echo "  make local-up       启动 Pulsar，并创建 pte_live_net"
	@echo "  make local-down     停止 Pulsar"
	@echo "  make deploy-up      部署启动 Pulsar"
	@echo "  make deploy-down    部署停止 Pulsar"
	@echo ""

local-up:
	$(COMPOSE_MQ) up -d

local-down:
	$(COMPOSE_MQ) down

local-restart:
	$(COMPOSE_MQ) restart

local-logs:
	$(COMPOSE_MQ) logs -f mq-pulsar

local-ps:
	$(COMPOSE_MQ) ps -a

# 部署命令与本地命令同构，便于脚本统一调用。
deploy-up: local-up
deploy-down: local-down
deploy-restart: local-restart
deploy-logs: local-logs
deploy-ps: local-ps
