-- MySQL migration script for Federation Backend Database
-- Generated from Go GORM models

SET FOREIGN_KEY_CHECKS = 0;

-- Create enums equivalents
-- Note: MySQL doesn't have true enums like PostgreSQL, so we use ENUM types

-- Create tables in dependency order

-- files table (referenced by multiple tables)
CREATE TABLE IF NOT EXISTS `files` (
                                       `id` INT AUTO_INCREMENT PRIMARY KEY,
                                       `name` VARCHAR(255) NOT NULL,
    `size` BIGINT NOT NULL,
    `path` VARCHAR(500) NOT NULL,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- chapters table
CREATE TABLE IF NOT EXISTS `chapters` (
                                          `id` INT AUTO_INCREMENT PRIMARY KEY,
                                          `name` VARCHAR(255) NOT NULL,
    `page` ENUM('news', 'gallery', 'documents') NOT NULL,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- teams table
CREATE TABLE IF NOT EXISTS `teams` (
                                       `id` INT AUTO_INCREMENT PRIMARY KEY,
                                       `team_name` VARCHAR(255) NOT NULL,
    `sex` ENUM('female', 'male') NOT NULL DEFAULT 'male',
    `team_logo_id` BIGINT UNSIGNED,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (`team_logo_id`) REFERENCES `files`(`id`) ON DELETE SET NULL
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- matches table
CREATE TABLE IF NOT EXISTS `matches` (
                                         `id` INT AUTO_INCREMENT PRIMARY KEY,
                                         `league` VARCHAR(255) NOT NULL,
    `date` TIMESTAMP NOT NULL,
    `sex` ENUM('female', 'male') NOT NULL,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- match_teams junction table (many-to-many relationship)
CREATE TABLE IF NOT EXISTS `match_teams` (
                                             `match_id` INT NOT NULL,
                                             `team_id` INT NOT NULL,
                                             `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                             PRIMARY KEY (`match_id`, `team_id`),
    FOREIGN KEY (`match_id`) REFERENCES `matches`(`id`) ON DELETE CASCADE,
    FOREIGN KEY (`team_id`) REFERENCES `teams`(`id`) ON DELETE CASCADE
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- news table
CREATE TABLE IF NOT EXISTS `news` (
                                      `id` INT AUTO_INCREMENT PRIMARY KEY,
                                      `heading` VARCHAR(500) NOT NULL,
    `description` TEXT NOT NULL,
    `date` TIMESTAMP NOT NULL,
    `chapter_id` INT UNSIGNED NOT NULL,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (`chapter_id`) REFERENCES `chapters`(`id`) ON DELETE CASCADE
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- news_images junction table
CREATE TABLE IF NOT EXISTS `news_images` (
                                             `news_id` INT NOT NULL,
                                             `file_id` INT NOT NULL,
                                             `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                             PRIMARY KEY (`news_id`, `file_id`),
    FOREIGN KEY (`news_id`) REFERENCES `news`(`id`) ON DELETE CASCADE,
    FOREIGN KEY (`file_id`) REFERENCES `files`(`id`) ON DELETE CASCADE
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- history_items table
CREATE TABLE IF NOT EXISTS `history_items` (
                                               `id` INT AUTO_INCREMENT PRIMARY KEY,
                                               `heading` VARCHAR(500) NOT NULL,
    `description` TEXT NOT NULL,
    `year` INT NOT NULL,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- history_item_images junction table (assuming similar structure to news_images)
CREATE TABLE IF NOT EXISTS `history_item_images` (
                                                     `history_item_id` INT NOT NULL,
                                                     `file_id` INT NOT NULL,
                                                     `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                                     PRIMARY KEY (`history_item_id`, `file_id`),
    FOREIGN KEY (`history_item_id`) REFERENCES `history_items`(`id`) ON DELETE CASCADE,
    FOREIGN KEY (`file_id`) REFERENCES `files`(`id`) ON DELETE CASCADE
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- gallery_items table
CREATE TABLE IF NOT EXISTS `gallery_items` (
                                               `id` INT AUTO_INCREMENT PRIMARY KEY,
                                               `chapter_id` INT UNSIGNED NOT NULL,
                                               `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                               `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                                               FOREIGN KEY (`chapter_id`) REFERENCES `chapters`(`id`) ON DELETE CASCADE
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- gallery_item_images junction table
CREATE TABLE IF NOT EXISTS `gallery_item_images` (
                                                     `gallery_item_id` INT NOT NULL,
                                                     `file_id` INT NOT NULL,
                                                     `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                                     PRIMARY KEY (`gallery_item_id`, `file_id`),
    FOREIGN KEY (`gallery_item_id`) REFERENCES `gallery_items`(`id`) ON DELETE CASCADE,
    FOREIGN KEY (`file_id`) REFERENCES `files`(`id`) ON DELETE CASCADE
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- documents table
CREATE TABLE IF NOT EXISTS `documents` (
                                           `id` INT AUTO_INCREMENT PRIMARY KEY,
                                           `name` VARCHAR(255) NOT NULL,
    `size` BIGINT NOT NULL,
    `path` VARCHAR(500) NOT NULL,
    `chapter` ENUM('rules', 'regulations') NOT NULL,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- call_backs table
CREATE TABLE IF NOT EXISTS `call_backs` (
                                            `id` INT AUTO_INCREMENT PRIMARY KEY,
                                            `name` VARCHAR(255) NOT NULL,
    `phone` VARCHAR(50) NOT NULL,
    `email` VARCHAR(255) NULL,
    `team_name` VARCHAR(255) NULL,
    `callback_type` ENUM('team_application', 'callback_request') NOT NULL DEFAULT 'callback_request',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- users table
CREATE TABLE IF NOT EXISTS `users` (
                                       `id` INT AUTO_INCREMENT PRIMARY KEY,
                                       `username` VARCHAR(255) NOT NULL UNIQUE,
    `password` VARCHAR(255) NOT NULL,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Create indexes for better performance
CREATE INDEX `idx_matches_date` ON `matches`(`date`);
CREATE INDEX `idx_matches_league` ON `matches`(`league`);
CREATE INDEX `idx_news_date` ON `news`(`date`);
CREATE INDEX `idx_news_chapter` ON `news`(`chapter_id`);
CREATE INDEX `idx_history_items_year` ON `history_items`(`year`);
CREATE INDEX `idx_teams_sex` ON `teams`(`sex`);
CREATE INDEX `idx_call_backs_type` ON `call_backs`(`callback_type`);

SET FOREIGN_KEY_CHECKS = 1;

-- Insert default chapters if needed
INSERT IGNORE INTO `chapters` (`name`, `page`) VALUES
('News Chapter', 'news'),
('Gallery Chapter', 'gallery'),
('Documents Chapter', 'documents');