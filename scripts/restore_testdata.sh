#!/bin/bash

# Скрипт для восстановления базы данных с тестовыми данными
# Usage: ./scripts/restore_testdata.sh [backup_file.sql]

set -e  # Exit on any error

BACKUP_FILE=""
DOCKER_CONTAINER="go_rest_tut_mysql"
DB_NAME="ecom"

# Функция для вывода цветного текста
print_info() {
    echo -e "\033[34m[INFO]\033[0m $1"
}

print_success() {
    echo -e "\033[32m[SUCCESS]\033[0m $1"
}

print_error() {
    echo -e "\033[31m[ERROR]\033[0m $1"
}

# Проверяем аргументы
if [ $# -eq 0 ]; then
    # Если файл не указан, ищем последний бэкап
    BACKUP_FILE=$(ls -t backups/ecom_testdata_*.sql 2>/dev/null | head -n1)
    if [ -z "$BACKUP_FILE" ]; then
        print_error "No backup files found in backups/ directory"
        echo "Usage: $0 [backup_file.sql]"
        exit 1
    fi
    print_info "Using latest backup: $BACKUP_FILE"
else
    BACKUP_FILE="$1"
    if [ ! -f "$BACKUP_FILE" ]; then
        print_error "Backup file '$BACKUP_FILE' not found"
        exit 1
    fi
fi

# Проверяем, что Docker контейнер запущен
if ! docker ps | grep -q "$DOCKER_CONTAINER"; then
    print_error "Docker container '$DOCKER_CONTAINER' is not running"
    print_info "Please start it with: docker-compose up -d"
    exit 1
fi

print_info "🔄 Starting database restoration..."

# 1. Останавливаем приложение (если запущено)
print_info "Checking if application is running..."
if lsof -i :8080 >/dev/null 2>&1; then
    print_info "⚠️  Application is running on port 8080. Please stop it first."
    echo "You can stop it with Ctrl+C if running in terminal"
    read -p "Press Enter when application is stopped..."
fi

# 2. Удаляем текущую базу данных
print_info "🗑️  Dropping existing database..."
docker exec "$DOCKER_CONTAINER" mysql -u root -ppassword -e "DROP DATABASE IF EXISTS $DB_NAME;" 2>/dev/null

# 3. Создаём новую базу данных
print_info "🏗️  Creating new database..."
docker exec "$DOCKER_CONTAINER" mysql -u root -ppassword -e "CREATE DATABASE $DB_NAME;" 2>/dev/null

# 4. Восстанавливаем данные из бэкапа
print_info "📥 Restoring data from backup..."
docker exec -i "$DOCKER_CONTAINER" mysql -u root -ppassword "$DB_NAME" < "$BACKUP_FILE"

# 5. Проверяем успешность восстановления
print_info "✅ Verifying restoration..."
TABLES_COUNT=$(docker exec "$DOCKER_CONTAINER" mysql -u root -ppassword -e "USE $DB_NAME; SHOW TABLES;" 2>/dev/null | wc -l)
USERS_COUNT=$(docker exec "$DOCKER_CONTAINER" mysql -u root -ppassword -e "USE $DB_NAME; SELECT COUNT(*) FROM users;" 2>/dev/null | tail -n1)
PRODUCTS_COUNT=$(docker exec "$DOCKER_CONTAINER" mysql -u root -ppassword -e "USE $DB_NAME; SELECT COUNT(*) FROM products;" 2>/dev/null | tail -n1)
MOVEMENTS_COUNT=$(docker exec "$DOCKER_CONTAINER" mysql -u root -ppassword -e "USE $DB_NAME; SELECT COUNT(*) FROM inventory_movements;" 2>/dev/null | tail -n1)

print_success "🎉 Database restoration completed!"
print_success "📊 Restored data summary:"
echo "   • Tables: $((TABLES_COUNT-1))"
echo "   • Users: $USERS_COUNT"
echo "   • Products: $PRODUCTS_COUNT"
echo "   • Inventory movements: $MOVEMENTS_COUNT"
echo ""
print_success "✅ You can now start your application with: make run"