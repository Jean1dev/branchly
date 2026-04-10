SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c

ROOT := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))

LOCAL_MONGODB_URI ?= mongodb://127.0.0.1:27017
LOCAL_RUNNER_URL ?= http://127.0.0.1:8081

.PHONY: help mongo-up dev

help:
	@echo "Alvos:"
	@echo "  make mongo-up  Garante MongoDB (serviço mongodb do docker compose)"
	@echo "  make dev       mongo-up + API + runner + web (pnpm dev)"
	@echo "Requisitos: Docker, Go, pnpm. Env: branchly-api/.env, branchly-runner/.env, branchly-web"

mongo-up:
	@cd "$(ROOT)" && \
	if [ -n "$$(docker compose ps mongodb --status running -q 2>/dev/null)" ]; then \
		echo "MongoDB (docker compose) já está em execução."; \
	elif docker ps --format '{{.Image}}' | grep -qE '^mongo(:|$$)'; then \
		echo "Já existe um container MongoDB em execução; não iniciando o serviço compose."; \
	else \
		echo "Iniciando MongoDB (docker compose)..."; \
		docker compose up -d mongodb; \
	fi

dev: mongo-up
	@cd "$(ROOT)" && \
	trap 'kill $$(jobs -p) 2>/dev/null; wait $$(jobs -p) 2>/dev/null || true' EXIT INT TERM; \
	cd "$(ROOT)/branchly-api" && MONGODB_URI="$(LOCAL_MONGODB_URI)" RUNNER_URL="$(LOCAL_RUNNER_URL)" go run ./cmd/api & \
	cd "$(ROOT)/branchly-runner" && MONGODB_URI="$(LOCAL_MONGODB_URI)" go run ./cmd/runner & \
	cd "$(ROOT)/branchly-web" && pnpm dev & \
	wait
