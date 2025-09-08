START TRANSACTION;

CREATE TABLE IF NOT EXISTS `provinces` (
    `id`         INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `name`       VARCHAR(255) NOT NULL UNIQUE,
    INDEX `idx_name` (`name`)
);

CREATE TABLE IF NOT EXISTS `cities` (
    `id`          INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `province_id` INT UNSIGNED NOT NULL,
    `name`        VARCHAR(255) NOT NULL UNIQUE,
    INDEX `idx_province_id` (`province_id`),
    INDEX `idx_name` (`name`),
    CONSTRAINT `fk_cities_province_id_provinces` FOREIGN KEY (`province_id`) REFERENCES `provinces`(`id`) ON DELETE RESTRICT
);

CREATE TABLE IF NOT EXISTS `districts` (
    `id`         INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `city_id`    INT UNSIGNED NOT NULL,
    `name`       VARCHAR(255) NOT NULL,
    INDEX `idx_city_id` (`city_id`),
    INDEX `idx_name` (`name`),
    CONSTRAINT `fk_districts_city_id_cities` FOREIGN KEY (`city_id`) REFERENCES `cities`(`id`) ON DELETE RESTRICT
);

CREATE TABLE IF NOT EXISTS `companies` (
    `id`                INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `admin_id`          INT UNSIGNED NOT NULL,
    `name`              VARCHAR(255) NOT NULL,
    `icon`              VARCHAR(255) DEFAULT NULL,
    `icon_updated_at`   TIMESTAMP    NULL DEFAULT NULL,
    `created_at`        TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`        TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at`        TIMESTAMP    NULL DEFAULT NULL,
    `name_active`       VARCHAR(255) GENERATED ALWAYS AS (IF(`deleted_at` IS NULL, `name`, NULL)) STORED,
    INDEX `idx_name` (`name`),
    INDEX `idx_admin_id` (`admin_id`),
    UNIQUE KEY `uq_companies_name_active` (`name_active`)
);

CREATE TABLE IF NOT EXISTS `users` (
    `id`              INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `company_id`      INT UNSIGNED NOT NULL,
    `username`        VARCHAR(255) NOT NULL,
    `email`           VARCHAR(255) NOT NULL,
    `password`        VARCHAR(255) NOT NULL,
    `fullname`        VARCHAR(255) NOT NULL,
    `created_at`      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at`      TIMESTAMP    NULL     DEFAULT NULL,
    `username_active` VARCHAR(255) GENERATED ALWAYS AS (IF(`deleted_at` IS NULL, `username`, NULL)) STORED,
    `email_active`    VARCHAR(255) GENERATED ALWAYS AS (IF(`deleted_at` IS NULL, `email`, NULL)) STORED,
    INDEX `idx_company_id` (`company_id`),
    INDEX `idx_username` (`username`),
    INDEX `idx_email` (`email`),
    INDEX `idx_fullname` (`fullname`),
    CONSTRAINT `fk_users_company_id_companies` FOREIGN KEY (`company_id`) REFERENCES `companies`(`id`) ON DELETE RESTRICT,
    UNIQUE KEY `uq_users_company_username_active` (`company_id`, `username_active`),
    UNIQUE KEY `uq_users_company_email_active` (`company_id`, `email_active`)
);

CREATE TABLE IF NOT EXISTS `clients` (
    `id`                INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `company_id`        INT UNSIGNED NOT NULL,
    `district_id`       INT UNSIGNED NOT NULL,
    `code`              VARCHAR(255) NOT NULL,
    `name`              VARCHAR(255) NOT NULL,
    `phone`             VARCHAR(15)  NOT NULL,
    `fax`               VARCHAR(50)  DEFAULT NULL,
    `icon`              VARCHAR(255) DEFAULT NULL,
    `icon_updated_at`   TIMESTAMP    NULL DEFAULT NULL,
    `pic_name`          VARCHAR(255) NOT NULL,
    `pic_phone`         VARCHAR(15)  NOT NULL,
    `village`           VARCHAR(100) NOT NULL,
    `postal_code`       VARCHAR(20)  NOT NULL,
    `address`           TEXT         NOT NULL,
    `created_at`        TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`        TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at`        TIMESTAMP    NULL DEFAULT NULL,
    `name_active`       VARCHAR(255) GENERATED ALWAYS AS (IF(`deleted_at` IS NULL, `name`, NULL)) STORED,
    `code_active`       VARCHAR(255) GENERATED ALWAYS AS (IF(`deleted_at` IS NULL, `code`, NULL)) STORED,
    INDEX `idx_company_id` (`company_id`),
    INDEX `idx_district_id` (`district_id`),
    INDEX `idx_code` (`code`),
    INDEX `idx_name` (`name`),
    INDEX `idx_pic_name` (`pic_name`),
    CONSTRAINT `fk_clients_company_id_companies` FOREIGN KEY (`company_id`) REFERENCES `companies`(`id`) ON DELETE RESTRICT,
    CONSTRAINT `fk_clients_district_id_districts` FOREIGN KEY (`district_id`) REFERENCES `districts`(`id`) ON DELETE RESTRICT,
    UNIQUE KEY `uq_clients_company_name_active` (`company_id`, `name_active`),
    UNIQUE KEY `uq_clients_company_code_active` (`company_id`, `code_active`)
);

CREATE TABLE IF NOT EXISTS `roles` (
    `id`          INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `code`        VARCHAR(255) NOT NULL,
    `name`        VARCHAR(255) NOT NULL,
    `super_admin` BOOLEAN      NOT NULL DEFAULT FALSE,
    `description` TEXT         DEFAULT NULL,
    `created_at`  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at`  TIMESTAMP    NULL     DEFAULT NULL,
    `code_active` VARCHAR(255) GENERATED ALWAYS AS (IF(`deleted_at` IS NULL, `code`, NULL)) STORED,
    `name_active` VARCHAR(255) GENERATED ALWAYS AS (IF(`deleted_at` IS NULL, `name`, NULL)) STORED,
    INDEX `idx_code` (`code`),
    INDEX `idx_name` (`name`),
    UNIQUE KEY `uq_roles_code_active` (`code_active`),
    UNIQUE KEY `uq_roles_name_active` (`name_active`)
);

CREATE TABLE IF NOT EXISTS `permissions` (
    `id`          INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `code`        VARCHAR(255) NOT NULL,
    `name`        VARCHAR(255) NOT NULL,
    `description` TEXT         DEFAULT NULL,
    `created_at`  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at`  TIMESTAMP    NULL     DEFAULT NULL,
    `code_active` VARCHAR(255) GENERATED ALWAYS AS (IF(`deleted_at` IS NULL, `code`, NULL)) STORED,
    `name_active` VARCHAR(255) GENERATED ALWAYS AS (IF(`deleted_at` IS NULL, `name`, NULL)) STORED,
    INDEX `idx_code` (`code`),
    INDEX `idx_name` (`name`),
    UNIQUE KEY `uq_permissions_code_active` (`code_active`),
    UNIQUE KEY `uq_permissions_name_active` (`name_active`)
);

CREATE TABLE IF NOT EXISTS `user_roles` (
    `user_id` INT UNSIGNED NOT NULL,
    `role_id` INT UNSIGNED NOT NULL,
    PRIMARY KEY (`user_id`, `role_id`),
    CONSTRAINT `fk_user_roles_user_id_users` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE RESTRICT,
    CONSTRAINT `fk_user_roles_role_id_roles` FOREIGN KEY (`role_id`) REFERENCES `roles`(`id`) ON DELETE RESTRICT
);

CREATE TABLE IF NOT EXISTS `role_permissions` (
    `role_id`       INT UNSIGNED NOT NULL,
    `permission_id` INT UNSIGNED NOT NULL,
    PRIMARY KEY (`role_id`, `permission_id`),
    CONSTRAINT `fk_role_permissions_role_id_roles` FOREIGN KEY (`role_id`) REFERENCES `roles`(`id`) ON DELETE RESTRICT,
    CONSTRAINT `fk_role_permissions_permission_id_permissions` FOREIGN KEY (`permission_id`) REFERENCES `permissions`(`id`) ON DELETE RESTRICT
);

CREATE TABLE IF NOT EXISTS `main_features` (
    `id`          INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `code`        VARCHAR(255) NOT NULL,
    `name`        VARCHAR(255) NOT NULL,
    `key`         VARCHAR(255) NOT NULL,
    `is_active`   BOOLEAN      NOT NULL DEFAULT TRUE,
    `created_at`  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at`  TIMESTAMP    NULL     DEFAULT NULL,
    `code_active` VARCHAR(255) GENERATED ALWAYS AS (IF(`deleted_at` IS NULL, `code`, NULL)) STORED,
    `name_active` VARCHAR(255) GENERATED ALWAYS AS (IF(`deleted_at` IS NULL, `name`, NULL)) STORED,
    `key_active`  VARCHAR(255) GENERATED ALWAYS AS (IF(`deleted_at` IS NULL, `key`,  NULL)) STORED,
    INDEX `idx_code` (`code`),
    INDEX `idx_name` (`name`),
    INDEX `idx_key` (`key`),
    UNIQUE KEY `uq_main_features_code_active` (`code_active`),
    UNIQUE KEY `uq_main_features_name_active` (`name_active`),
    UNIQUE KEY `uq_main_features_key_active` (`key_active`)
);

CREATE TABLE IF NOT EXISTS `client_main_features` (
    `client_id`       INT UNSIGNED NOT NULL,
    `main_feature_id` INT UNSIGNED NOT NULL,
    `order`           INT          NOT NULL,
    PRIMARY KEY (`client_id`, `main_feature_id`),
    CONSTRAINT `fk_climf_client_id_clients` FOREIGN KEY (`client_id`) REFERENCES `clients`(`id`) ON DELETE RESTRICT,
    CONSTRAINT `fk_climf_main_feature_id_main_features` FOREIGN KEY (`main_feature_id`) REFERENCES `main_features`(`id`) ON DELETE RESTRICT
);

CREATE TABLE IF NOT EXISTS `support_features` (
    `id`          INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `code`        VARCHAR(255) NOT NULL,
    `name`        VARCHAR(255) NOT NULL,
    `key`         VARCHAR(255) NOT NULL,
    `is_active`   BOOLEAN      NOT NULL DEFAULT TRUE,
    `created_at`  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at`  TIMESTAMP    NULL     DEFAULT NULL,
    `code_active` VARCHAR(255) GENERATED ALWAYS AS (IF(`deleted_at` IS NULL, `code`, NULL)) STORED,
    `name_active` VARCHAR(255) GENERATED ALWAYS AS (IF(`deleted_at` IS NULL, `name`, NULL)) STORED,
    `key_active`  VARCHAR(255) GENERATED ALWAYS AS (IF(`deleted_at` IS NULL, `key`,  NULL)) STORED,
    INDEX `idx_code` (`code`),
    INDEX `idx_name` (`name`),
    INDEX `idx_key` (`key`),
    UNIQUE KEY `uq_support_features_code_active` (`code_active`),
    UNIQUE KEY `uq_support_features_name_active` (`name_active`),
    UNIQUE KEY `uq_support_features_key_active` (`key_active`)
);

CREATE TABLE IF NOT EXISTS `client_support_features` (
    `client_id`          INT UNSIGNED NOT NULL,
    `support_feature_id` INT UNSIGNED NOT NULL,
    `order`              INT          NOT NULL,
    PRIMARY KEY (`client_id`, `support_feature_id`),
    CONSTRAINT `fk_clisupf_client_id_clients` FOREIGN KEY (`client_id`) REFERENCES `clients`(`id`) ON DELETE RESTRICT,
    CONSTRAINT `fk_clisupf_support_feature_id_support_features` FOREIGN KEY (`support_feature_id`) REFERENCES `support_features`(`id`) ON DELETE RESTRICT
);

COMMIT;