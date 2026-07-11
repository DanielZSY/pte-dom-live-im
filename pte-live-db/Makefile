SHELL := /bin/bash
.DEFAULT_GOAL := help

DB_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
ROOT_DIR := $(abspath $(DB_DIR)/..)
COMPOSE_DB := docker compose -p pte_live_db -f $(DB_DIR)/docker-compose.yaml
MYSQL_DATABASE ?= pte_live_im
MYSQL_ROOT_PASSWORD ?= pte_live_root

.PHONY: help local-up local-down local-restart local-logs local-ps local-sql-load local-reset deploy-up deploy-down deploy-restart deploy-logs deploy-ps deploy-sql-load deploy-reset

help:
	@echo ""
	@echo "pte-live-db"
	@echo "  make local-up          启动 MySQL + Redis"
	@echo "  make local-sql-load    导入 ../pte-live-sql/001_im_schema.sql"
	@echo "  make local-down        停止 DB 服务"
	@echo "  make deploy-up         部署启动 MySQL + Redis"
	@echo "  make deploy-sql-load   部署导入 IM SQL"
	@echo "  make deploy-down       停止 DB 服务"
	@echo ""

local-up:
	$(COMPOSE_DB) up -d

local-down:
	$(COMPOSE_DB) down

local-restart:
	$(COMPOSE_DB) restart

local-logs:
	$(COMPOSE_DB) logs -f

local-ps:
	$(COMPOSE_DB) ps -a

local-sql-load:
	$(COMPOSE_DB) exec -T db-mysql mysql -uroot -p$(MYSQL_ROOT_PASSWORD) $(MYSQL_DATABASE) < $(ROOT_DIR)/pte-live-sql/001_im_schema.sql

local-reset:
	@echo "Refusing to reset database automatically. Run: docker volume rm pte_live_db_mysql_data pte_live_cache_redis_data"

# 部署命令与本地命令同构，便于脚本统一调用。
deploy-up: local-up
deploy-down: local-down
deploy-restart: local-restart
deploy-logs: local-logs
deploy-ps: local-ps
deploy-sql-load: local-sql-load
deploy-reset: local-reset
