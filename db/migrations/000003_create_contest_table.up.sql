CREATE TABLE contests(
    contest_id BIGINT AUTO_INCREMENT primary key,
    title VARCHAR(255) NOT NULL,
    description VARCHAR(255),
    max_team_count INT,
    total_point INT DEFAULT 100,
    contest_type VARCHAR(16) NOT NULL,
    contest_status VARCHAR(16) NOT NULL,
    started_at DATETIME,
    ended_at DATETIME,
    auto_start BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    modified_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT chk_contest_type
        CHECK (contest_type IN ('TOURNAMENT', 'LEAGUE', 'CASUAL')),
    CONSTRAINT chk_contest_status
        CHECK (contest_status IN ('PENDING', 'ACTIVE', 'FINISHED', 'CANCELLED')),
    CONSTRAINT chk_contest_dates
        CHECK (ended_at IS NULL OR started_at < ended_at),

    INDEX idx_contest_status (contest_status),
    INDEX idx_contest_type (contest_type),
    INDEX idx_contest_dates (started_at, ended_at),
    INDEX idx_auto_start_status (auto_start, contest_status, started_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

