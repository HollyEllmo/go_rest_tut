-- MySQL dump 10.13  Distrib 8.0.43, for Linux (aarch64)
--
-- Host: localhost    Database: ecom
-- ------------------------------------------------------
-- Server version	8.0.43

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!50503 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `inventory_movements`
--

DROP TABLE IF EXISTS `inventory_movements`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `inventory_movements` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `product_id` int unsigned NOT NULL,
  `movement_type` enum('IN','OUT') NOT NULL,
  `quantity` int unsigned NOT NULL,
  `reason` varchar(100) NOT NULL,
  `reference_id` int unsigned DEFAULT NULL,
  `reference_type` enum('ORDER','RESTOCK','ADJUSTMENT','RETURN') DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_product_id` (`product_id`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_reference` (`reference_type`,`reference_id`),
  CONSTRAINT `inventory_movements_ibfk_1` FOREIGN KEY (`product_id`) REFERENCES `products` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=22 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `inventory_movements`
--

LOCK TABLES `inventory_movements` WRITE;
/*!40000 ALTER TABLE `inventory_movements` DISABLE KEYS */;
INSERT INTO `inventory_movements` VALUES (1,1,'IN',50,'Initial stock - iPhone 15 Pro',NULL,'RESTOCK','2025-07-28 06:36:54'),(2,2,'IN',25,'Initial stock - MacBook Air M3',NULL,'RESTOCK','2025-07-28 06:36:54'),(3,3,'IN',100,'Initial stock - AirPods Pro 2',NULL,'RESTOCK','2025-07-28 06:36:54'),(4,4,'IN',40,'Initial stock - iPad Air',NULL,'RESTOCK','2025-07-28 06:36:54'),(5,5,'IN',75,'Initial stock - Apple Watch Series 9',NULL,'RESTOCK','2025-07-28 06:36:54'),(6,6,'IN',35,'Initial stock - Samsung Galaxy S24',NULL,'RESTOCK','2025-07-28 06:36:54'),(7,7,'IN',60,'Initial stock - Sony WH-1000XM5',NULL,'RESTOCK','2025-07-28 06:36:54'),(8,8,'IN',80,'Initial stock - Nintendo Switch OLED',NULL,'RESTOCK','2025-07-28 06:36:54'),(9,9,'IN',20,'Initial stock - Dell XPS 13',NULL,'RESTOCK','2025-07-28 06:36:54'),(10,10,'IN',45,'Initial stock - Google Pixel 8',NULL,'RESTOCK','2025-07-28 06:36:54'),(11,1,'OUT',1,'Sold to customer',1,'ORDER','2025-07-28 06:37:12'),(12,3,'OUT',1,'Sold to customer',1,'ORDER','2025-07-28 06:37:12'),(13,2,'OUT',1,'Reserved for order',2,'ORDER','2025-07-28 06:37:24'),(14,4,'OUT',1,'Reserved for order',2,'ORDER','2025-07-28 06:37:24'),(15,5,'OUT',2,'Sold to customer',3,'ORDER','2025-07-28 06:37:55'),(16,7,'OUT',1,'Sold to customer',3,'ORDER','2025-07-28 06:37:55'),(17,10,'OUT',1,'Sold to customer',3,'ORDER','2025-07-28 06:37:55'),(18,8,'OUT',1,'Reserved for order',4,'ORDER','2025-07-28 06:42:35'),(19,1,'IN',20,'Restock iPhone 15 Pro',NULL,'RESTOCK','2025-07-28 06:42:35'),(20,3,'IN',50,'Restock AirPods Pro 2',NULL,'RESTOCK','2025-07-28 06:42:35'),(21,7,'IN',1,'Customer return - defective unit',NULL,'RETURN','2025-07-28 06:42:35');
/*!40000 ALTER TABLE `inventory_movements` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `order_items`
--

DROP TABLE IF EXISTS `order_items`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `order_items` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `orderId` int unsigned NOT NULL,
  `productId` int unsigned NOT NULL,
  `quantity` int NOT NULL,
  `price` decimal(10,2) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `orderId` (`orderId`),
  KEY `productId` (`productId`),
  CONSTRAINT `order_items_ibfk_1` FOREIGN KEY (`orderId`) REFERENCES `orders` (`id`),
  CONSTRAINT `order_items_ibfk_2` FOREIGN KEY (`productId`) REFERENCES `products` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=9 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `order_items`
--

LOCK TABLES `order_items` WRITE;
/*!40000 ALTER TABLE `order_items` DISABLE KEYS */;
INSERT INTO `order_items` VALUES (1,1,1,1,999.99),(2,1,3,1,249.99),(3,2,2,1,1299.99),(4,2,4,1,599.99),(5,3,5,2,399.99),(6,3,7,1,349.99),(7,3,10,1,699.99),(8,4,8,1,349.99);
/*!40000 ALTER TABLE `order_items` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `orders`
--

DROP TABLE IF EXISTS `orders`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `orders` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `userId` int unsigned NOT NULL,
  `total` decimal(10,2) NOT NULL,
  `status` enum('pending','completed','cancelled') NOT NULL DEFAULT 'pending',
  `address` text NOT NULL,
  `createdAt` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `userId` (`userId`),
  CONSTRAINT `orders_ibfk_1` FOREIGN KEY (`userId`) REFERENCES `users` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `orders`
--

LOCK TABLES `orders` WRITE;
/*!40000 ALTER TABLE `orders` DISABLE KEYS */;
INSERT INTO `orders` VALUES (1,1,1249.98,'completed','123 Main St, New York','2025-07-28 06:37:12'),(2,2,1899.98,'pending','456 Oak Ave, Los Angeles','2025-07-28 06:37:24'),(3,3,1749.96,'completed','789 Pine St, Chicago','2025-07-28 06:37:55'),(4,4,349.99,'pending','321 Elm St, Miami','2025-07-28 06:42:35');
/*!40000 ALTER TABLE `orders` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `products`
--

DROP TABLE IF EXISTS `products`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `products` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL,
  `description` text NOT NULL,
  `image` varchar(255) NOT NULL,
  `price` decimal(10,2) NOT NULL,
  `createdAt` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=11 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `products`
--

LOCK TABLES `products` WRITE;
/*!40000 ALTER TABLE `products` DISABLE KEYS */;
INSERT INTO `products` VALUES (1,'iPhone 15 Pro','Latest Apple smartphone with A17 Pro chip','iphone15pro.jpg',999.99,'2025-07-28 06:35:23'),(2,'MacBook Air M3','Ultra-thin laptop with M3 chip and 13-inch display','macbookair.jpg',1299.99,'2025-07-28 06:35:23'),(3,'AirPods Pro 2','Premium wireless earbuds with active noise cancellation','airpods.jpg',249.99,'2025-07-28 06:35:23'),(4,'iPad Air','Versatile tablet with M1 chip and 10.9-inch display','ipadair.jpg',599.99,'2025-07-28 06:35:23'),(5,'Apple Watch Series 9','Advanced smartwatch with health monitoring','applewatch.jpg',399.99,'2025-07-28 06:35:23'),(6,'Samsung Galaxy S24','Android flagship with AI features','galaxys24.jpg',899.99,'2025-07-28 06:35:23'),(7,'Sony WH-1000XM5','Premium noise-canceling headphones','sonywh1000xm5.jpg',349.99,'2025-07-28 06:35:23'),(8,'Nintendo Switch OLED','Portable gaming console with OLED screen','switcholed.jpg',349.99,'2025-07-28 06:35:23'),(9,'Dell XPS 13','Premium Windows laptop with Intel processors','dellxps13.jpg',1199.99,'2025-07-28 06:35:23'),(10,'Google Pixel 8','Android phone with pure Google experience','pixel8.jpg',699.99,'2025-07-28 06:35:23');
/*!40000 ALTER TABLE `products` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `schema_migrations`
--

DROP TABLE IF EXISTS `schema_migrations`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `schema_migrations` (
  `version` bigint NOT NULL,
  `dirty` tinyint(1) NOT NULL,
  PRIMARY KEY (`version`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `schema_migrations`
--

LOCK TABLES `schema_migrations` WRITE;
/*!40000 ALTER TABLE `schema_migrations` DISABLE KEYS */;
INSERT INTO `schema_migrations` VALUES (20250726103948,0);
/*!40000 ALTER TABLE `schema_migrations` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `users`
--

DROP TABLE IF EXISTS `users`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `users` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `firstName` varchar(255) NOT NULL,
  `lastName` varchar(255) NOT NULL,
  `email` varchar(255) NOT NULL,
  `password` varchar(255) NOT NULL,
  `createdAt` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `email` (`email`),
  UNIQUE KEY `email_2` (`email`)
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `users`
--

LOCK TABLES `users` WRITE;
/*!40000 ALTER TABLE `users` DISABLE KEYS */;
INSERT INTO `users` VALUES (1,'ergrwg','wefwef','valid@gmail.com','$2a$10$lsqKiTWSQ8DGC6rkv0SrXu4yIrTrODZzmL37RZf8jNky9vrKP1qgi','2025-07-26 05:27:38'),(2,'John','Doe','john@example.com','$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi','2025-07-28 06:35:01'),(3,'Jane','Smith','jane@example.com','$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi','2025-07-28 06:35:01'),(4,'Mike','Johnson','mike@example.com','$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi','2025-07-28 06:35:01'),(5,'Sarah','Wilson','sarah@example.com','$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi','2025-07-28 06:35:01');
/*!40000 ALTER TABLE `users` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Dumping routines for database 'ecom'
--
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2025-07-28  6:48:11
