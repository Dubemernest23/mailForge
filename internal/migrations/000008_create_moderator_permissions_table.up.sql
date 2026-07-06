CREATE TABLE moderator_permissions (
    id            BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    public_id     VARCHAR(36) NOT NULL UNIQUE,
    moderator_id  BIGINT UNSIGNED NOT NULL,
    permission    VARCHAR(100) NOT NULL,
    granted_by    BIGINT UNSIGNED NOT NULL,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (moderator_id) REFERENCES users(id),
    FOREIGN KEY (granted_by)   REFERENCES users(id),
    
    UNIQUE KEY uq_mod_permission (moderator_id, permission),
    INDEX idx_permission (permission)

) ENGINE=InnoDB
    DEFAULT CHARSET=utf8mb4
    COLLATE=utf8mb4_unicode_ci;