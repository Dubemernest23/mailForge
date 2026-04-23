CREATE TABLE subscribers (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    public_id CHAR(36) NOT NULL UNIQUE,
    user_id BIGINT UNSIGNED NOT NULL,
    email VARCHAR(255) NOT NULL,
    name VARCHAR(255) NULL,
    status ENUM('active', 'unsubscribed', 'bounced', 'complained') 
        NOT NULL DEFAULT 'active',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_subscribers_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE,
    
    -- indexes
    UNIQUE KEY uniq_user_email (user_id, email),
    -- INDEX idx_subscribers_user_id (user_id),
    INDEX idx_subscribers_email (email)

) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;
