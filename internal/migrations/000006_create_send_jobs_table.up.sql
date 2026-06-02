CREATE TABLE send_jobs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    public_id CHAR(36) NOT NULL,
    job_type ENUM(
        'welcome_email',
        'password_reset_email',
        'csv_import',
        'campaign_send'
    ) NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    campaign_id BIGINT UNSIGNED NULL,
    subscriber_id BIGINT UNSIGNED NULL,
    payload JSON NOT NULL,
    status ENUM('pending', 'processing', 'delivered', 'failed') NOT NULL DEFAULT 'pending',
    attempts TINYINT UNSIGNED NOT NULL DEFAULT 0,
    last_error TEXT NULL,
    scheduled_at DATETIME NULL,
    processed_at DATETIME NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uniq_send_jobs_public_id (public_id),
    INDEX idx_send_jobs_status (status),
    INDEX idx_send_jobs_job_type (job_type),
    INDEX idx_send_jobs_user_id (user_id),
    INDEX idx_send_jobs_campaign_id (campaign_id),

    CONSTRAINT fk_send_jobs_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_send_jobs_campaign FOREIGN KEY (campaign_id) REFERENCES campaigns(id) ON DELETE SET NULL

) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;