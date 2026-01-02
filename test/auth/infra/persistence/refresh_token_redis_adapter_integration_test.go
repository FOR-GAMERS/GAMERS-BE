package persistence

import (
	"GAMERS-BE/internal/auth/domain"
	authcommand "GAMERS-BE/internal/auth/infra/persistence/command"
	authquery "GAMERS-BE/internal/auth/infra/persistence/query"
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type redisTestFixture struct {
	ctx            context.Context
	client         *redis.Client
	commandAdapter *authcommand.RefreshTokenRedisCommandAdapter
	queryAdapter   *authquery.RefreshTokenRedisQueryAdapter
}

func setupRedisTest(t *testing.T) *redisTestFixture {
	ctx := context.Background()

	// Connect to Redis
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       2, // Use DB 2 for integration tests
	})

	// Ping to ensure Redis is available
	err := client.Ping(ctx).Err()
	require.NoError(t, err, "Redis must be running for integration tests. Start Redis with: docker run -d -p 6379:6379 redis")

	// Clear all keys in test DB
	err = client.FlushDB(ctx).Err()
	require.NoError(t, err)

	commandAdapter := authcommand.NewRefreshTokenRedisCommandAdapter(client)
	queryAdapter := authquery.NewRefreshTokenRedisQueryAdapter(client)

	return &redisTestFixture{
		ctx:            ctx,
		client:         client,
		commandAdapter: commandAdapter,
		queryAdapter:   queryAdapter,
	}
}

func (f *redisTestFixture) cleanup() {
	f.client.FlushDB(f.ctx).Err()
	f.client.Close()
}

func TestRedisIntegration_SaveAndFind(t *testing.T) {
	fixture := setupRedisTest(t)
	defer fixture.cleanup()

	t.Run("Save and Find Refresh Token", func(t *testing.T) {
		// Given
		token := "test-refresh-token-123"
		userID := int64(456)
		expiresAt := time.Now().Add(7 * 24 * time.Hour).Unix()
		refreshToken := domain.NewRefreshToken(token, userID, expiresAt)
		ttl := 7 * 24 * time.Hour

		// When - Save
		err := fixture.commandAdapter.Save(fixture.ctx, refreshToken, ttl)

		// Then
		assert.NoError(t, err)

		// When - Find
		found, err := fixture.queryAdapter.FindByToken(fixture.ctx, token)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, token, found.Token)
		assert.Equal(t, userID, found.UserID)
		assert.Equal(t, expiresAt, found.ExpiresAt)
	})
}

func TestRedisIntegration_SaveWithTTL(t *testing.T) {
	fixture := setupRedisTest(t)
	defer fixture.cleanup()

	t.Run("Token Should Expire After TTL", func(t *testing.T) {
		// Given
		token := "expiring-token"
		userID := int64(789)
		expiresAt := time.Now().Add(5 * time.Second).Unix()
		refreshToken := domain.NewRefreshToken(token, userID, expiresAt)
		ttl := 2 * time.Second

		// When - Save with short TTL
		err := fixture.commandAdapter.Save(fixture.ctx, refreshToken, ttl)
		assert.NoError(t, err)

		// Then - Token should exist immediately
		found, err := fixture.queryAdapter.FindByToken(fixture.ctx, token)
		assert.NoError(t, err)
		assert.NotNil(t, found)

		// Wait for TTL to expire
		time.Sleep(3 * time.Second)

		// Then - Token should not exist after TTL
		found, err = fixture.queryAdapter.FindByToken(fixture.ctx, token)
		assert.Error(t, err)
		assert.Nil(t, found)
	})
}

func TestRedisIntegration_Delete(t *testing.T) {
	fixture := setupRedisTest(t)
	defer fixture.cleanup()

	t.Run("Delete Existing Token", func(t *testing.T) {
		// Given
		token := "token-to-delete"
		userID := int64(101)
		expiresAt := time.Now().Add(24 * time.Hour).Unix()
		refreshToken := domain.NewRefreshToken(token, userID, expiresAt)
		ttl := 24 * time.Hour

		err := fixture.commandAdapter.Save(fixture.ctx, refreshToken, ttl)
		assert.NoError(t, err)

		// When - Delete
		err = fixture.commandAdapter.Delete(fixture.ctx, token)

		// Then
		assert.NoError(t, err)

		// Verify token is deleted
		found, err := fixture.queryAdapter.FindByToken(fixture.ctx, token)
		assert.Error(t, err)
		assert.Nil(t, found)
	})

	t.Run("Delete Non-existent Token Should Not Error", func(t *testing.T) {
		// When
		err := fixture.commandAdapter.Delete(fixture.ctx, "non-existent-token")

		// Then - Should not return error
		assert.NoError(t, err)
	})
}

func TestRedisIntegration_DeleteByUserID(t *testing.T) {
	fixture := setupRedisTest(t)
	defer fixture.cleanup()

	t.Run("Delete All Tokens For User", func(t *testing.T) {
		// Given - Multiple tokens for same user
		userID := int64(202)
		tokens := []string{"token1", "token2", "token3"}
		expiresAt := time.Now().Add(24 * time.Hour).Unix()
		ttl := 24 * time.Hour

		for _, token := range tokens {
			refreshToken := domain.NewRefreshToken(token, userID, expiresAt)
			err := fixture.commandAdapter.Save(fixture.ctx, refreshToken, ttl)
			assert.NoError(t, err)
		}

		// When - Delete by user ID
		err := fixture.commandAdapter.DeleteByUserID(fixture.ctx, uint(userID))

		// Then
		assert.NoError(t, err)

		// Verify all tokens are deleted
		for _, token := range tokens {
			found, err := fixture.queryAdapter.FindByToken(fixture.ctx, token)
			assert.Error(t, err)
			assert.Nil(t, found)
		}
	})

	t.Run("Delete By UserID Should Not Affect Other Users", func(t *testing.T) {
		// Given - Tokens for different users
		user1ID := int64(301)
		user2ID := int64(302)
		token1 := "user1-token"
		token2 := "user2-token"
		expiresAt := time.Now().Add(24 * time.Hour).Unix()
		ttl := 24 * time.Hour

		refreshToken1 := domain.NewRefreshToken(token1, user1ID, expiresAt)
		refreshToken2 := domain.NewRefreshToken(token2, user2ID, expiresAt)

		err := fixture.commandAdapter.Save(fixture.ctx, refreshToken1, ttl)
		assert.NoError(t, err)
		err = fixture.commandAdapter.Save(fixture.ctx, refreshToken2, ttl)
		assert.NoError(t, err)

		// When - Delete user1's tokens
		err = fixture.commandAdapter.DeleteByUserID(fixture.ctx, uint(user1ID))
		assert.NoError(t, err)

		// Then - User1's token should be deleted
		found, err := fixture.queryAdapter.FindByToken(fixture.ctx, token1)
		assert.Error(t, err)
		assert.Nil(t, found)

		// Then - User2's token should still exist
		found, err = fixture.queryAdapter.FindByToken(fixture.ctx, token2)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, token2, found.Token)
	})
}

func TestRedisIntegration_ExistsByToken(t *testing.T) {
	fixture := setupRedisTest(t)
	defer fixture.cleanup()

	t.Run("Token Exists", func(t *testing.T) {
		// Given
		token := "existing-token"
		userID := int64(401)
		expiresAt := time.Now().Add(24 * time.Hour).Unix()
		refreshToken := domain.NewRefreshToken(token, userID, expiresAt)
		ttl := 24 * time.Hour

		err := fixture.commandAdapter.Save(fixture.ctx, refreshToken, ttl)
		assert.NoError(t, err)

		// When
		exists, err := fixture.queryAdapter.ExistsByToken(fixture.ctx, token)

		// Then
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("Token Does Not Exist", func(t *testing.T) {
		// When
		exists, err := fixture.queryAdapter.ExistsByToken(fixture.ctx, "non-existent-token")

		// Then
		assert.NoError(t, err)
		assert.False(t, exists)
	})
}

func TestRedisIntegration_MultipleTokensForSameUser(t *testing.T) {
	fixture := setupRedisTest(t)
	defer fixture.cleanup()

	t.Run("User Can Have Multiple Active Tokens", func(t *testing.T) {
		// Given - Same user with multiple tokens (e.g., multiple devices)
		userID := int64(501)
		token1 := "device1-token"
		token2 := "device2-token"
		token3 := "device3-token"
		expiresAt := time.Now().Add(24 * time.Hour).Unix()
		ttl := 24 * time.Hour

		tokens := []string{token1, token2, token3}
		for _, token := range tokens {
			refreshToken := domain.NewRefreshToken(token, userID, expiresAt)
			err := fixture.commandAdapter.Save(fixture.ctx, refreshToken, ttl)
			assert.NoError(t, err)
		}

		// Then - All tokens should be retrievable
		for _, token := range tokens {
			found, err := fixture.queryAdapter.FindByToken(fixture.ctx, token)
			assert.NoError(t, err)
			assert.NotNil(t, found)
			assert.Equal(t, userID, found.UserID)
		}
	})
}

func TestRedisIntegration_UpdateToken(t *testing.T) {
	fixture := setupRedisTest(t)
	defer fixture.cleanup()

	t.Run("Overwrite Existing Token", func(t *testing.T) {
		// Given - Initial token
		token := "same-token-key"
		userID1 := int64(601)
		userID2 := int64(602)
		expiresAt1 := time.Now().Add(24 * time.Hour).Unix()
		expiresAt2 := time.Now().Add(48 * time.Hour).Unix()
		ttl := 24 * time.Hour

		refreshToken1 := domain.NewRefreshToken(token, userID1, expiresAt1)
		err := fixture.commandAdapter.Save(fixture.ctx, refreshToken1, ttl)
		assert.NoError(t, err)

		// When - Save again with same token key but different data
		refreshToken2 := domain.NewRefreshToken(token, userID2, expiresAt2)
		err = fixture.commandAdapter.Save(fixture.ctx, refreshToken2, ttl)
		assert.NoError(t, err)

		// Then - Should have the new data
		found, err := fixture.queryAdapter.FindByToken(fixture.ctx, token)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, userID2, found.UserID)
		assert.Equal(t, expiresAt2, found.ExpiresAt)
	})
}

func TestRedisIntegration_ConcurrentOperations(t *testing.T) {
	fixture := setupRedisTest(t)
	defer fixture.cleanup()

	t.Run("Concurrent Save Operations", func(t *testing.T) {
		// Given
		userID := int64(701)
		expiresAt := time.Now().Add(24 * time.Hour).Unix()
		ttl := 24 * time.Hour
		numTokens := 10

		// When - Save multiple tokens concurrently
		done := make(chan bool, numTokens)
		for i := 0; i < numTokens; i++ {
			go func(index int) {
				token := "concurrent-token-" + string(rune('0'+index))
				refreshToken := domain.NewRefreshToken(token, userID, expiresAt)
				err := fixture.commandAdapter.Save(fixture.ctx, refreshToken, ttl)
				assert.NoError(t, err)
				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < numTokens; i++ {
			<-done
		}

		// Then - All tokens should be saved
		for i := 0; i < numTokens; i++ {
			token := "concurrent-token-" + string(rune('0'+i))
			found, err := fixture.queryAdapter.FindByToken(fixture.ctx, token)
			assert.NoError(t, err)
			assert.NotNil(t, found)
		}
	})
}

func TestRedisIntegration_EdgeCases(t *testing.T) {
	fixture := setupRedisTest(t)
	defer fixture.cleanup()

	t.Run("Very Long Token String", func(t *testing.T) {
		// Given - Very long token
		longToken := ""
		for i := 0; i < 1000; i++ {
			longToken += "a"
		}
		userID := int64(801)
		expiresAt := time.Now().Add(24 * time.Hour).Unix()
		refreshToken := domain.NewRefreshToken(longToken, userID, expiresAt)
		ttl := 24 * time.Hour

		// When
		err := fixture.commandAdapter.Save(fixture.ctx, refreshToken, ttl)

		// Then
		assert.NoError(t, err)

		found, err := fixture.queryAdapter.FindByToken(fixture.ctx, longToken)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, longToken, found.Token)
	})

	t.Run("Special Characters In Token", func(t *testing.T) {
		// Given - Token with special characters
		specialToken := "token-with-special!@#$%^&*()_+-=[]{}|;:',.<>?/~`"
		userID := int64(802)
		expiresAt := time.Now().Add(24 * time.Hour).Unix()
		refreshToken := domain.NewRefreshToken(specialToken, userID, expiresAt)
		ttl := 24 * time.Hour

		// When
		err := fixture.commandAdapter.Save(fixture.ctx, refreshToken, ttl)

		// Then
		assert.NoError(t, err)

		found, err := fixture.queryAdapter.FindByToken(fixture.ctx, specialToken)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, specialToken, found.Token)
	})

	t.Run("Zero UserID", func(t *testing.T) {
		// Given
		token := "zero-userid-token"
		userID := int64(0)
		expiresAt := time.Now().Add(24 * time.Hour).Unix()
		refreshToken := domain.NewRefreshToken(token, userID, expiresAt)
		ttl := 24 * time.Hour

		// When
		err := fixture.commandAdapter.Save(fixture.ctx, refreshToken, ttl)

		// Then
		assert.NoError(t, err)

		found, err := fixture.queryAdapter.FindByToken(fixture.ctx, token)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, userID, found.UserID)
	})
}
