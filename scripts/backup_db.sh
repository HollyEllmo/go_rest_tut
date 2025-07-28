#!/bin/bash

# Скрипт для создания бэкапа базы данных
# Usage: ./scripts/backup_db.sh [custom_name]

set -e  # Exit on any error

DOCKER_CONTAINER="go_rest_tut_mysql"
DB_NAME="ecom"
BACKUP_DIR="backups"

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

# Создаём директорию для бэкапов если её нет
mkdir -p "$BACKUP_DIR"

# Определяем имя файла бэкапа
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
if [ $# -eq 0 ]; then
    BACKUP_FILE="$BACKUP_DIR/ecom_backup_$TIMESTAMP.sql"
else
    CUSTOM_NAME="$1"
    BACKUP_FILE="$BACKUP_DIR/ecom_${CUSTOM_NAME}_$TIMESTAMP.sql"
fi

# Проверяем, что Docker контейнер запущен
if ! docker ps | grep -q "$DOCKER_CONTAINER"; then
    print_error "Docker container '$DOCKER_CONTAINER' is not running"
    print_info "Please start it with: docker-compose up -d"
    exit 1
fi

print_info "🗄️  Creating database backup..."
print_info "📁 Backup file: $BACKUP_FILE"

# Создаём бэкап
docker exec "$DOCKER_CONTAINER" mysqldump \
    -u root -ppassword \
    --routines \
    --triggers \
    --single-transaction \
    --add-drop-database \
    --databases "$DB_NAME" > "$BACKUP_FILE" 2>/dev/null

# Проверяем размер файла
BACKUP_SIZE=$(du -h "$BACKUP_FILE" | cut -f1)
print_success "✅ Backup created successfully!"
print_success "📦 File: $BACKUP_FILE ($BACKUP_SIZE)"

# Показываем статистику
TABLES_COUNT=$(docker exec "$DOCKER_CONTAINER" mysql -u root -ppassword -e "USE $DB_NAME; SHOW TABLES;" 2>/dev/null | wc -l)
USERS_COUNT=$(docker exec "$DOCKER_CONTAINER" mysql -u root -ppassword -e "USE $DB_NAME; SELECT COUNT(*) FROM users;" 2>/dev/null | tail -n1)
PRODUCTS_COUNT=$(docker exec "$DOCKER_CONTAINER" mysql -u root -ppassword -e "USE $DB_NAME; SELECT COUNT(*) FROM products;" 2>/dev/null | tail -n1)

print_info "📊 Backed up data:"
echo "   • Tables: $((TABLES_COUNT-1))"
echo "   • Users: $USERS_COUNT"
echo "   • Products: $PRODUCTS_COUNT"

# Показываем последние 5 бэкапов
print_info "📋 Recent backups:"
ls -lt "$BACKUP_DIR"/*.sql 2>/dev/null | head -5 | while read line; do
    echo "   $line"
done