#!/bin/bash

# Скрипт для настройки тестовой базы данных
# Usage: ./scripts/setup_test_db.sh

set -e

DOCKER_CONTAINER="go_rest_tut_mysql"
TEST_DB_NAME="ecom_test"

print_info() {
    echo -e "\033[34m[INFO]\033[0m $1"
}

print_success() {
    echo -e "\033[32m[SUCCESS]\033[0m $1"
}

print_error() {
    echo -e "\033[31m[ERROR]\033[0m $1"
}

# Проверяем, что Docker контейнер запущен
if ! docker ps | grep -q "$DOCKER_CONTAINER"; then
    print_error "Docker container '$DOCKER_CONTAINER' is not running"
    print_info "Please start it with: docker-compose up -d"
    exit 1
fi

print_info "🏗️  Setting up test database..."

# Создаём тестовую базу данных
print_info "Creating test database '$TEST_DB_NAME'..."
docker exec "$DOCKER_CONTAINER" mysql -u root -ppassword -e "
    DROP DATABASE IF EXISTS $TEST_DB_NAME;
    CREATE DATABASE $TEST_DB_NAME;
" 2>/dev/null

# Применяем миграции к тестовой базе
print_info "Applying migrations to test database..."
export DB_NAME="$TEST_DB_NAME"
go run cmd/migrate/main.go up

print_success "✅ Test database '$TEST_DB_NAME' is ready!"
print_info "You can now run tests with: go test ./cmd/service/inventory -v"