# Testing Guide

This document describes how to run and understand the concurrent safety tests for the inventory system.

## Prerequisites

1. **Docker MySQL container running:**
   ```bash
   docker-compose up -d
   ```

2. **Test database setup:**
   ```bash
   ./scripts/setup_test_db.sh
   ```

## Test Overview

The tests verify that the inventory system correctly handles concurrent stock reservations without race conditions or overselling.

### Test Types

#### 1. `TestConcurrentStockReservation`
**Purpose:** Verify that concurrent goroutines can't oversell products.

**Scenario:**
- Initial stock: 10 items
- 20 goroutines try to reserve 1 item each simultaneously
- Expected: Exactly 10 successful reservations, 10 failures

**What it tests:**
- `FOR UPDATE` row-level locking works correctly
- No race conditions in read-modify-write operations
- Atomic stock reservations

#### 2. `TestOversellePrevention`
**Purpose:** Verify that single requests can't exceed available stock.

**Scenario:**
- Initial stock: 5 items
- Try to reserve 6 items
- Expected: Error with "insufficient stock" message

#### 3. `TestHighLoadConcurrentReservations`
**Purpose:** Stress test with multiple products and high concurrency.

**Scenario:**
- 3 products with different stock levels (100, 50, 10)
- 200 concurrent goroutines reserving 1-3 items randomly
- Expected: No negative stock, proper error handling

## Running Tests

### All Inventory Tests
```bash
go test ./cmd/service/inventory -v
```

### Specific Test
```bash
go test ./cmd/service/inventory -v -run TestConcurrentStockReservation
```

### Skip Long-Running Tests
```bash
go test ./cmd/service/inventory -v -short
```

### With Race Detection
```bash
go test ./cmd/service/inventory -v -race
```

## Expected Output

### Successful Test Run
```
=== RUN   TestConcurrentStockReservation
    store_test.go:123: Successful reservations: 10
    store_test.go:124: Failed reservations: 10
    store_test.go:125: Expected successful: 10
    store_test.go:145: Insufficient stock errors: 10
--- PASS: TestConcurrentStockReservation (0.15s)

=== RUN   TestOversellePrevention
--- PASS: TestOversellePrevention (0.02s)

=== RUN   TestHighLoadConcurrentReservations
    store_test.go:252: High-load test results:
    store_test.go:253: Total successful reservations: 85
    store_test.go:254: Total failed reservations: 115
    store_test.go:255: Total goroutines: 200
    store_test.go:266: Product 11 final stock: 22 (started with 100)
    store_test.go:266: Product 12 final stock: 18 (started with 50)
    store_test.go:266: Product 13 final stock: 0 (started with 10)
--- PASS: TestHighLoadConcurrentReservations (0.45s)
```

## Understanding the Results

### Key Metrics to Watch

1. **Exact Reservation Count:**
   - Successful + Failed should equal Total Goroutines
   - Successful should never exceed available stock

2. **Final Stock Levels:**
   - Must never be negative
   - Should equal (Initial - Reserved)

3. **Error Messages:**
   - Failed reservations should have proper "insufficient stock" errors
   - Error messages should show correct available quantities

### What Indicates a Race Condition

❌ **Bad signs:**
- More successful reservations than initial stock
- Negative final stock
- Inconsistent error messages
- Test failures or panics

✅ **Good signs:**
- Exact math: Initial Stock = Final Stock + Successful Reservations
- All failed requests have proper error messages
- No negative stock levels
- Consistent behavior across multiple runs

## Debugging Failed Tests

### If Tests Fail

1. **Check database connection:**
   ```bash
   docker ps | grep mysql
   ```

2. **Verify test database exists:**
   ```bash
   docker exec go_rest_tut_mysql mysql -u root -ppassword -e "SHOW DATABASES;"
   ```

3. **Clean test database:**
   ```bash
   ./scripts/setup_test_db.sh
   ```

4. **Run with verbose logging:**
   ```bash
   go test ./cmd/service/inventory -v -run TestConcurrentStockReservation
   ```

### Common Issues

- **Database not running:** Start with `docker-compose up -d`
- **Test database missing:** Run `./scripts/setup_test_db.sh`
- **Concurrent failures:** This indicates a race condition bug
- **Wrong stock math:** Check the inventory ledger logic

## Performance Benchmarks

Expected performance on modern hardware:
- **Concurrent test:** ~100-200ms for 20 goroutines
- **High-load test:** ~400-600ms for 200 goroutines
- **Memory usage:** <10MB during tests

Slower performance may indicate:
- Database connection issues
- Lock contention problems
- Inefficient queries

## Integration with CI/CD

Add to your CI pipeline:
```yaml
- name: Setup Test Database
  run: ./scripts/setup_test_db.sh

- name: Run Concurrent Safety Tests
  run: go test ./cmd/service/inventory -v -race

- name: Run Short Tests Only
  run: go test ./cmd/service/inventory -v -short
```