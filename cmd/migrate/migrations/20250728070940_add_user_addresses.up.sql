CREATE TABLE user_addresses (
  `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` INT UNSIGNED NOT NULL,
  `title` VARCHAR(100) NOT NULL, -- e.g., "Home", "Work", "Office"
  `first_name` VARCHAR(100) NOT NULL,
  `last_name` VARCHAR(100) NOT NULL,
  `company` VARCHAR(100) NULL,
  `address_line_1` VARCHAR(255) NOT NULL,
  `address_line_2` VARCHAR(255) NULL,
  `city` VARCHAR(100) NOT NULL,
  `state_province` VARCHAR(100) NOT NULL,
  `postal_code` VARCHAR(20) NOT NULL,
  `country` VARCHAR(100) NOT NULL DEFAULT 'United States',
  `phone` VARCHAR(20) NULL,
  `is_default` BOOLEAN NOT NULL DEFAULT FALSE,
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  INDEX idx_user_id (user_id),
  INDEX idx_user_default (user_id, is_default),
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);