CREATE TABLE list_subscribers (
    list_id BIGINT UNSIGNED NOT NULL,
    subscriber_id BIGINT UNSIGNED NOT NULL,
    status ENUM('subscribed', 'unsubscribed') 
        NOT NULL DEFAULT 'subscribed',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (list_id, subscriber_id),


    CONSTRAINT fk_ls_list
        FOREIGN KEY (list_id)
        REFERENCES lists(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_ls_subscriber
        FOREIGN KEY (subscriber_id)
        REFERENCES subscribers(id)
        ON DELETE CASCADE,

    -- Index for reverse lookup (subscriber → lists)
    INDEX idx_ls_subscriber_id (subscriber_id)

) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;
