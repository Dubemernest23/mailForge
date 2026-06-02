CREATE TABLE dead_letter_jobs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    original_job_id BIGINT UNSIGNED NOT NULL,
    job_type ENUM(
        'welcome_email',
        'password_reset_email',
        'csv_import',
        'campaign_send'
    ) NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    campaign_id BIGINT UNSIGNED NULL,
    subscriber_id BIGINT UNSIGNED NULL,
    failure_reason TEXT NOT NULL,
    payload JSON NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_dlj_original_job_id (original_job_id),
    INDEX idx_dlj_job_type (job_type),
    INDEX idx_dlj_user_id (user_id),

    CONSTRAINT fk_dlj_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE

) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;