#!/usr/bin/env sh
set -eu

missing=0

check_required() {
	name="$1"
	version_command="$2"

	if command -v "$name" >/dev/null 2>&1; then
		printf "Found required tool: %s - " "$name"
		sh -c "$version_command" 2>/dev/null || printf "version unavailable\n"
	else
		printf "Missing required tool: %s\n" "$name"
		missing=1
	fi
}

check_optional() {
	name="$1"
	version_command="$2"

	if command -v "$name" >/dev/null 2>&1; then
		printf "Found optional tool: %s - " "$name"
		sh -c "$version_command" 2>/dev/null || printf "version unavailable\n"
	else
		printf "Optional tool not found: %s\n" "$name"
	fi
}

check_docker_compose() {
	if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then
		printf "Found required tool: docker compose - "
		docker compose version
	else
		printf "Missing required tool: docker compose\n"
		missing=1
	fi
}

check_required "go" "go version"
check_required "node" "node --version"
check_required "npm" "npm --version"
check_required "docker" "docker --version"
check_docker_compose

check_optional "migrate" "migrate -version"
check_optional "sqlc" "sqlc version"

if [ "$missing" -ne 0 ]; then
	printf "One or more required tools are missing. Install them and retry.\n"
	exit 1
fi

printf "All required tools are available.\n"
