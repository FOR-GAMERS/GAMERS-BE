-- Create main_banners table for homepage banners
CREATE TABLE main_banners (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    image_key VARCHAR(512) NOT NULL,
    title VARCHAR(255),
    link_url VARCHAR(512),
    display_order INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    modified_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- Create index for active banners ordered by display_order
CREATE INDEX idx_main_banners_active_order ON main_banners(is_active, display_order);
