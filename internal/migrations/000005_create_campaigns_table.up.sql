CREATE TABLE campaigns (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    public_id CHAR(36) NOT NULL UNIQUE,
    list_id BIGINT UNSIGNED NULL,
    name VARCHAR(255) NOT NULL,
    subject VARCHAR(255) NOT NULL,
    preview_text VARCHAR(255) NOT NULL,
    body LONGTEXT NOT NULL,
    status ENUM('draft', 'scheduled', 'sending', 'sent', 'cancelled')
        NOT NULL DEFAULT 'draft',
    scheduled_at DATETIME NULL,
    started_at DATETIME NULL,
    completed_at DATETIME NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    -- Constraints
    CONSTRAINT fk_campaigns_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_campaigns_list
        FOREIGN KEY (list_id)
        REFERENCES lists(id)
        ON DELETE SET NULL,

    -- indexes
    UNIQUE KEY uniq_user_campaign_name (user_id, name),
    INDEX idx_campaigns_user_id (user_id),
    INDEX idx_campaigns_list_id (list_id),
    INDEX idx_campaigns_status (status),
    INDEX idx_campaigns_scheduled_at (scheduled_at)

) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;
