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
    INDEX idx_contest_dates (started_at, ended_at)

) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE contests_members(
    user_id BIGINT NOT NULL,
    contest_id BIGINT NOT NULL,
    member_type VARCHAR(16) NOT NULL,
    leader_type VARCHAR(8) NOT NULL,
    point INT DEFAULT 0,
    PRIMARY KEY (user_id, contest_id),

    CONSTRAINT chk_member_type
        CHECK (member_type IN ('STAFF', 'NORMAL')),
    CONSTRAINT chk_leader_type
        CHECK (leader_type IN ('LEADER', 'MEMBER')),
    CONSTRAINT chk_point_range
        CHECK (point >= 0),

    FOREIGN KEY (user_id) REFERENCES users(user_id)
        ON DELETE CASCADE
        ON UPDATE CASCADE,
    FOREIGN KEY (contest_id) REFERENCES contests(contest_id)
        ON DELETE CASCADE
        ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;