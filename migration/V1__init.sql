CREATE TABLE users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    email VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,

    UNIQUE KEY unique_users_email (email),
    INDEX idx_users_created_at (created_at)
);

CREATE TABLE profiles (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    username VARCHAR(16) NOT NULL,
    tag VARCHAR(5)  NOT NULL,
    avatar VARCHAR(255),
    bio VARCHAR(255),
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,

    UNIQUE KEY unique_username_tag (username, tag),
    INDEX idx_user_id (user_id),

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);