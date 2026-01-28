package persistence_test

import (
	"GAMERS-BE/internal/contest/infra/persistence/adapter"
	"GAMERS-BE/internal/global/utils"
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"GAMERS-BE/test/global/support"

	"github.com/redis/go-redis/v9"
)

// seedApplicationData - 벤치마크용 Redis 신청 데이터 세팅
func seedApplicationData(ctx context.Context, client *redis.Client, contestId int64, memberCount int) {
	pipe := client.Pipeline()
	for i := 0; i < memberCount; i++ {
		uid := int64(i + 1)
		appKey := utils.GetApplicationKey(contestId, uid)
		data, _ := json.Marshal(map[string]interface{}{
			"user_id":    uid,
			"contest_id": contestId,
			"status":     "PENDING",
		})
		pipe.Set(ctx, appKey, data, time.Hour)
		pipe.ZAdd(ctx, utils.GetPendingKey(contestId), redis.Z{
			Score:  float64(time.Now().Unix()),
			Member: uid,
		})
	}
	pipe.Expire(ctx, utils.GetPendingKey(contestId), time.Hour)
	pipe.Exec(ctx)
}

// cleanApplicationData - 벤치마크용 Redis 데이터 정리
func cleanApplicationData(ctx context.Context, client *redis.Client, contestId int64, memberCount int) {
	pipe := client.Pipeline()
	for i := 0; i < memberCount; i++ {
		pipe.Del(ctx, utils.GetApplicationKey(contestId, int64(i+1)))
	}
	pipe.Del(ctx, utils.GetPendingKey(contestId))
	pipe.Del(ctx, utils.GetAcceptedKey(contestId))
	pipe.Del(ctx, utils.GetRejectedKey(contestId))
	pipe.Exec(ctx)
}

// seedNoiseKeys - SCAN 성능에 영향을 주는 노이즈 키 삽입
func seedNoiseKeys(ctx context.Context, client *redis.Client, count int) {
	const batchSize = 1000
	for start := 0; start < count; start += batchSize {
		pipe := client.Pipeline()
		end := start + batchSize
		if end > count {
			end = count
		}
		for i := start; i < end; i++ {
			pipe.Set(ctx, fmt.Sprintf("noise:key:%d", i), "x", time.Hour)
		}
		pipe.Exec(ctx)
	}
}

// cleanNoiseKeys - 노이즈 키 정리
func cleanNoiseKeys(ctx context.Context, client *redis.Client, count int) {
	const batchSize = 1000
	for start := 0; start < count; start += batchSize {
		pipe := client.Pipeline()
		end := start + batchSize
		if end > count {
			end = count
		}
		for i := start; i < end; i++ {
			pipe.Del(ctx, fmt.Sprintf("noise:key:%d", i))
		}
		pipe.Exec(ctx)
	}
}

// extendTTLByScan - 변경 전 방식: SCAN 패턴 매칭으로 키 탐색 후 TTL 연장
func extendTTLByScan(ctx context.Context, client *redis.Client, contestId int64, newTTL time.Duration) error {
	pattern := utils.GetContestPatternKey(contestId)
	var cursor uint64
	for {
		keys, nextCursor, err := client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}

		if len(keys) > 0 {
			pipe := client.Pipeline()
			for _, key := range keys {
				pipe.Expire(ctx, key, newTTL)
			}
			if _, err := pipe.Exec(ctx); err != nil {
				return err
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	return nil
}

// clearApplicationsByScan - 변경 전 방식: SCAN 패턴 매칭으로 키 탐색 후 삭제
func clearApplicationsByScan(ctx context.Context, client *redis.Client, contestId int64) error {
	pattern := utils.GetContestPatternKey(contestId)
	var cursor uint64
	for {
		keys, nextCursor, err := client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}

		if len(keys) > 0 {
			pipe := client.Pipeline()
			for _, key := range keys {
				pipe.Del(ctx, key)
			}
			if _, err := pipe.Exec(ctx); err != nil {
				return err
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	return nil
}

// BenchmarkExtendTTL_ScanVsDirect 는 변경 전(SCAN)과 변경 후(Set 직접 조회)의
// ExtendTTL 성능을 노이즈 키 수와 멤버 수 조합별로 비교합니다.
//
// SCAN은 전체 keyspace 크기에 비례하므로, 노이즈 키가 많을수록 느려집니다.
// Direct 방식은 해당 contest 멤버 수에만 비례하므로 노이즈에 영향받지 않습니다.
//
// 실행:
//
//	go test -bench=BenchmarkExtendTTL_ScanVsDirect -benchmem -count=3 -timeout=10m \
//	  ./test/contest/infra/persistence/...
func BenchmarkExtendTTL_ScanVsDirect(b *testing.B) {
	ctx := context.Background()
	redisContainer, err := support.SetupRedisContainer(ctx)
	if err != nil {
		b.Fatal(err)
	}
	defer redisContainer.Teardown(ctx)
	client := redisContainer.GetClient()

	contestId := int64(1)
	memberCounts := []int{50, 100}
	noiseCounts := []int{0, 10000, 100000}

	for _, noise := range noiseCounts {
		if noise > 0 {
			b.Logf("Seeding %d noise keys...", noise)
			seedNoiseKeys(ctx, client, noise)
		}

		for _, members := range memberCounts {
			label := fmt.Sprintf("noise=%d/members=%d", noise, members)

			// 변경 전: SCAN 방식
			b.Run("SCAN/"+label, func(b *testing.B) {
				seedApplicationData(ctx, client, contestId, members)
				defer cleanApplicationData(ctx, client, contestId, members)

				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					extendTTLByScan(ctx, client, contestId, time.Hour)
				}
			})

			// 변경 후: Direct 방식 (collectAllKeys 사용)
			b.Run("Direct/"+label, func(b *testing.B) {
				seedApplicationData(ctx, client, contestId, members)
				defer cleanApplicationData(ctx, client, contestId, members)

				redisAdapter := adapter.NewContestApplicationRedisAdapter(client)

				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					redisAdapter.ExtendTTL(ctx, contestId, time.Hour)
				}
			})
		}

		if noise > 0 {
			cleanNoiseKeys(ctx, client, noise)
		}
	}
}

// BenchmarkClearApplications_ScanVsDirect 는 변경 전(SCAN)과 변경 후(Set 직접 조회)의
// ClearApplications 성능을 비교합니다.
//
// ClearApplications는 키를 삭제하므로 매 반복마다 데이터를 다시 세팅해야 합니다.
// 따라서 setup 비용을 b.StopTimer/StartTimer로 제외합니다.
//
// 실행:
//
//	go test -bench=BenchmarkClearApplications_ScanVsDirect -benchmem -count=3 -timeout=10m \
//	  ./test/contest/infra/persistence/...
func BenchmarkClearApplications_ScanVsDirect(b *testing.B) {
	ctx := context.Background()
	redisContainer, err := support.SetupRedisContainer(ctx)
	if err != nil {
		b.Fatal(err)
	}
	defer redisContainer.Teardown(ctx)
	client := redisContainer.GetClient()

	contestId := int64(1)
	memberCount := 100
	noiseCount := 50000

	b.Logf("Seeding %d noise keys...", noiseCount)
	seedNoiseKeys(ctx, client, noiseCount)
	defer cleanNoiseKeys(ctx, client, noiseCount)

	// 변경 전: SCAN 방식
	b.Run(fmt.Sprintf("SCAN/members=%d", memberCount), func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			seedApplicationData(ctx, client, contestId, memberCount)
			b.StartTimer()

			clearApplicationsByScan(ctx, client, contestId)
		}
	})

	// 변경 후: Direct 방식
	b.Run(fmt.Sprintf("Direct/members=%d", memberCount), func(b *testing.B) {
		redisAdapter := adapter.NewContestApplicationRedisAdapter(client)

		for i := 0; i < b.N; i++ {
			b.StopTimer()
			seedApplicationData(ctx, client, contestId, memberCount)
			b.StartTimer()

			redisAdapter.ClearApplications(ctx, contestId)
		}
	})
}
