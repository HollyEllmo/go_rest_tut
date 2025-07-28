CREATE TABLE inventory_movements (
  `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
  `product_id` INT UNSIGNED NOT NULL,
  `movement_type` ENUM('IN', 'OUT') NOT NULL,
  `quantity` INT UNSIGNED NOT NULL,
  `reason` VARCHAR(100) NOT NULL,
  `reference_id` INT UNSIGNED NULL, -- for linking to orders/restocks
  `reference_type` ENUM('ORDER', 'RESTOCK', 'ADJUSTMENT', 'RETURN') NULL,
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  INDEX idx_product_id (product_id),
  INDEX idx_created_at (created_at),
  INDEX idx_reference (reference_type, reference_id),
  FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
);