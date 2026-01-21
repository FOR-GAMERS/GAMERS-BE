-- Create team_members table (team membership)
-- This table stores the relationship between teams and users
CREATE TABLE team_members (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    team_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    member_type VARCHAR(16) NOT NULL,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- Foreign key constraints
    CONSTRAINT fk_team_members_team
        FOREIGN KEY (team_id) REFERENCES teams(team_id)
        ON DELETE CASCADE,
    CONSTRAINT fk_team_members_user
        FOREIGN KEY (user_id) REFERENCES users(id)
        ON DELETE CASCADE,

    -- Unique constraint to prevent a user from joining the same team twice
    CONSTRAINT uq_team_members_team_user UNIQUE (team_id, user_id),

    -- Check constraint for member_type
    CONSTRAINT chk_team_member_type
        CHECK (member_type IN ('MEMBER', 'LEADER')),

    -- Indexes
    INDEX idx_team_members_team_id (team_id),
    INDEX idx_team_members_user_id (user_id),
    INDEX idx_team_members_member_type (member_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
