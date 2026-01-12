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

    INDEX idx_contest_members_contest_id (contest_id),
    INDEX idx_contest_members_type (member_type, leader_type),

    FOREIGN KEY (user_id) REFERENCES users(id)
        ON DELETE CASCADE
        ON UPDATE CASCADE,
    FOREIGN KEY (contest_id) REFERENCES contests(contest_id)
        ON DELETE CASCADE
        ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;