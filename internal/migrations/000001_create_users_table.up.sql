CREATE TABLE users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    public_id CHAR(36) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    email_verified_at DATETIME NULL,
    verification_token VARCHAR(255) NULL,
    verification_token_expires_at DATETIME NULL,
    is_verified BOOLEAN NOT NULL DEFAULT FALSE,
    role ENUM('user', 'admin', 'super_admin') NOT NULL DEFAULT 'user',
    status ENUM('active', 'suspended', 'deleted') NOT NULL DEFAULT 'active',
    last_login_at DATETIME NULL,
    failed_login_attempts INT UNSIGNED NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uniq_users_email (email),
    UNIQUE KEY uniq_users_verification_token (verification_token)

) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;
