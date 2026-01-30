package main

import "fmt"

var seedDataStatements []string

func init() {
	seedDataStatements = generateSeedData()
}

func generateSeedData() []string {
	stmts := []string{}

	// Create users (200 users for team members)
	for i := 1; i <= 200; i++ {
		stmts = append(stmts, fmt.Sprintf(
			`INSERT INTO users (id, email, password, username, tag, role) VALUES (%d, 'user%d@test.com', 'hashed_pw', 'User%d', '%04d', 'USER')`,
			i, i, i, i,
		))
	}

	// Contest 100: 10 teams x 5 members (finalized, FINISHED status)
	stmts = append(stmts, `INSERT INTO contests (contest_id, title, description, max_team_count, contest_type, contest_status, total_team_member) VALUES (100, 'Contest-10Teams', 'Load test contest with 10 teams', 10, 'LEAGUE', 'FINISHED', 5)`)
	stmts = append(stmts, generateTeamsAndMembers(100, 10, 5, 1)...)

	// Contest 101: 20 teams x 5 members
	stmts = append(stmts, `INSERT INTO contests (contest_id, title, description, max_team_count, contest_type, contest_status, total_team_member) VALUES (101, 'Contest-20Teams', 'Load test contest with 20 teams', 20, 'LEAGUE', 'FINISHED', 5)`)
	stmts = append(stmts, generateTeamsAndMembers(101, 20, 5, 51)...)

	// Contest 102: 50 teams x 5 members
	stmts = append(stmts, `INSERT INTO contests (contest_id, title, description, max_team_count, contest_type, contest_status, total_team_member) VALUES (102, 'Contest-50Teams', 'Load test contest with 50 teams', 50, 'LEAGUE', 'FINISHED', 5)`)
	stmts = append(stmts, generateTeamsAndMembers(102, 50, 5, 151)...)

	return stmts
}

func generateTeamsAndMembers(contestID, teamCount, membersPerTeam, teamIDStart int) []string {
	stmts := []string{}
	userID := 1

	for t := 0; t < teamCount; t++ {
		teamID := teamIDStart + t
		stmts = append(stmts, fmt.Sprintf(
			`INSERT INTO teams (team_id, contest_id, team_name) VALUES (%d, %d, 'Team-%d-%d')`,
			teamID, contestID, contestID, t+1,
		))

		for m := 0; m < membersPerTeam; m++ {
			memberType := "MEMBER"
			if m == 0 {
				memberType = "LEADER"
			}
			stmts = append(stmts, fmt.Sprintf(
				`INSERT INTO team_members (team_id, user_id, member_type) VALUES (%d, %d, '%s')`,
				teamID, userID, memberType,
			))
			userID++
			if userID > 200 {
				userID = 1
			}
		}
	}

	return stmts
}

var createTableStatements = []string{
	// 1. users
	`CREATE TABLE users (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		email VARCHAR(64) NOT NULL,
		password VARCHAR(255) NOT NULL,
		username VARCHAR(16) NOT NULL,
		tag VARCHAR(6) NOT NULL,
		bio VARCHAR(255),
		avatar TEXT,
		profile_key VARCHAR(512),
		role VARCHAR(16) NOT NULL DEFAULT 'USER',
		riot_name VARCHAR(32),
		riot_tag VARCHAR(8),
		region VARCHAR(10),
		current_tier INT,
		current_tier_patched VARCHAR(32),
		elo INT,
		ranking_in_tier INT,
		peak_tier INT,
		peak_tier_patched VARCHAR(32),
		valorant_updated_at TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		modified_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	)`,

	// 2. discord_accounts
	`CREATE TABLE discord_accounts (
		discord_id VARCHAR(255) PRIMARY KEY,
		user_id BIGINT NOT NULL,
		discord_avatar VARCHAR(255),
		discord_verified BOOLEAN DEFAULT FALSE
	)`,

	// 3. contests
	`CREATE TABLE contests (
		contest_id BIGINT AUTO_INCREMENT PRIMARY KEY,
		title VARCHAR(255) NOT NULL,
		description TEXT,
		max_team_count INT,
		total_point INT DEFAULT 100,
		contest_type VARCHAR(16) NOT NULL,
		contest_status VARCHAR(16) NOT NULL,
		started_at DATETIME,
		ended_at DATETIME,
		auto_start BOOLEAN DEFAULT FALSE,
		game_type VARCHAR(32),
		game_point_table_id BIGINT,
		total_team_member INT NOT NULL DEFAULT 5,
		discord_guild_id VARCHAR(255),
		discord_text_channel_id VARCHAR(255),
		thumbnail VARCHAR(512),
		banner_key VARCHAR(512),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		modified_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	)`,

	// 4. contests_members
	`CREATE TABLE contests_members (
		user_id BIGINT NOT NULL,
		contest_id BIGINT NOT NULL,
		member_type VARCHAR(16) NOT NULL,
		leader_type VARCHAR(8) NOT NULL,
		point INT DEFAULT 0,
		PRIMARY KEY (user_id, contest_id)
	)`,

	// 5. teams
	`CREATE TABLE teams (
		team_id BIGINT AUTO_INCREMENT PRIMARY KEY,
		contest_id BIGINT NOT NULL,
		team_name VARCHAR(50) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		modified_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	)`,

	// 6. team_members
	`CREATE TABLE team_members (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		team_id BIGINT NOT NULL,
		user_id BIGINT NOT NULL,
		member_type VARCHAR(16) NOT NULL,
		joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`,

	// 7. games
	`CREATE TABLE games (
		game_id BIGINT AUTO_INCREMENT PRIMARY KEY,
		contest_id BIGINT NOT NULL,
		game_status VARCHAR(16) NOT NULL,
		game_team_type VARCHAR(16) NOT NULL,
		started_at DATETIME,
		ended_at DATETIME,
		round INT,
		match_number INT,
		next_game_id BIGINT,
		bracket_position INT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		modified_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	)`,

	// 8. game_teams
	`CREATE TABLE game_teams (
		game_team_id BIGINT AUTO_INCREMENT PRIMARY KEY,
		game_id BIGINT NOT NULL,
		team_id BIGINT NOT NULL,
		grade INT
	)`,

	// 9. contest_comments
	`CREATE TABLE contest_comments (
		comment_id BIGINT AUTO_INCREMENT PRIMARY KEY,
		contest_id BIGINT NOT NULL,
		user_id BIGINT NOT NULL,
		content VARCHAR(255) NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		modified_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	)`,

	// 10. valorant_score_tables
	`CREATE TABLE valorant_score_tables (
		score_table_id BIGINT AUTO_INCREMENT PRIMARY KEY,
		radiant INT NOT NULL,
		immortal_3 INT NOT NULL DEFAULT 0,
		immortal_2 INT NOT NULL DEFAULT 0,
		immortal_1 INT NOT NULL DEFAULT 0,
		ascendant_3 INT NOT NULL DEFAULT 0,
		ascendant_2 INT NOT NULL DEFAULT 0,
		ascendant_1 INT NOT NULL DEFAULT 0,
		diamond_3 INT NOT NULL DEFAULT 0,
		diamond_2 INT NOT NULL DEFAULT 0,
		diamond_1 INT NOT NULL DEFAULT 0,
		platinum_3 INT NOT NULL DEFAULT 0,
		platinum_2 INT NOT NULL DEFAULT 0,
		platinum_1 INT NOT NULL DEFAULT 0,
		gold_3 INT NOT NULL DEFAULT 0,
		gold_2 INT NOT NULL DEFAULT 0,
		gold_1 INT NOT NULL DEFAULT 0,
		silver_3 INT NOT NULL DEFAULT 0,
		silver_2 INT NOT NULL DEFAULT 0,
		silver_1 INT NOT NULL DEFAULT 0,
		bronze_3 INT NOT NULL DEFAULT 0,
		bronze_2 INT NOT NULL DEFAULT 0,
		bronze_1 INT NOT NULL DEFAULT 0,
		iron_3 INT NOT NULL DEFAULT 0,
		iron_2 INT NOT NULL DEFAULT 0,
		iron_1 INT NOT NULL DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		modified_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	)`,

	// 11. main_banners
	`CREATE TABLE main_banners (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		image_key VARCHAR(512) NOT NULL,
		title VARCHAR(255),
		link_url VARCHAR(512),
		display_order INT NOT NULL DEFAULT 0,
		is_active BOOLEAN NOT NULL DEFAULT TRUE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		modified_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	)`,

	// 12. notifications
	`CREATE TABLE notifications (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		user_id BIGINT NOT NULL,
		type VARCHAR(50) NOT NULL,
		title VARCHAR(255) NOT NULL,
		message TEXT,
		data JSON,
		is_read BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`,
}
