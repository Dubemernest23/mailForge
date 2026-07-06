CREATE TABLE tracking_events (
    id            BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    public_id     VARCHAR(36) NOT NULL UNIQUE,
    campaign_id   BIGINT UNSIGNED NOT NULL,
    subscriber_id BIGINT UNSIGNED NOT NULL,
    event_type    ENUM('open', 'click', 'unsubscribe') NOT NULL,
    metadata      JSON,
    ip_address    VARCHAR(45),
    user_agent    TEXT,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_tracking_events_subscriber FOREIGN KEY (subscriber_id) REFERENCES subscribers(id) ON DELETE RESTRICT,
    CONSTRAINT fk_tracking_events_campaign FOREIGN KEY (campaign_id) REFERENCES campaigns(id) ON DELETE CASCADE,
    
    UNIQUE KEY uniq_user_tracking_pub_id (public_id),

    INDEX idx_tracking_events_campaign_id (campaign_id),
    INDEX idx_tracking_events_subscriber_id (subscriber_id),
    INDEX idx_tracking_events_event_type (event_type),
    INDEX idx_created_at (created_at)

)  ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;