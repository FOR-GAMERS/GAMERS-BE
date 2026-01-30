package persistence_test

import (
	"github.com/FOR-GAMERS/GAMERS-BE/internal/contest/domain"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/contest/infra/persistence/adapter"
	"github.com/FOR-GAMERS/GAMERS-BE/test/global/support"
	"context"
	"fmt"
	"testing"
	"time"

	"gorm.io/gorm"
)

// saveBatchLoop - 변경 전 방식: 트랜잭션 내에서 건건이 Create
func saveBatchLoop(db *gorm.DB, members []*domain.ContestMember) error {
	return db.Transaction(func(tx *gorm.DB) error {
		for _, member := range members {
			if err := tx.Create(member).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// saveBatchBulk - 변경 후 방식: CreateInBatches
func saveBatchBulk(db *gorm.DB, members []*domain.ContestMember) error {
	return db.CreateInBatches(members, 100).Error
}

// createBenchContest - 벤치마크용 대회 생성 헬퍼
func createBenchContest(contestAdapter *adapter.ContestDatabaseAdapter, label string) *domain.Contest {
	contest := &domain.Contest{
		Title:           label,
		Description:     "benchmark test",
		MaxTeamCount:    8,
		TotalPoint:      100,
		ContestType:     domain.ContestTypeTournament,
		ContestStatus:   domain.ContestStatusPending,
		StartedAt:       time.Now(),
		EndedAt:         time.Now().Add(48 * time.Hour),
		TotalTeamMember: 5,
	}
	saved, _ := contestAdapter.Save(contest)
	return saved
}

// makeMembers - 벤치마크용 멤버 슬라이스 생성 헬퍼
func makeMembers(count int, contestID int64, offset int) []*domain.ContestMember {
	members := make([]*domain.ContestMember, count)
	for j := 0; j < count; j++ {
		members[j] = domain.NewContestMember(
			int64(offset+j+1), contestID,
			domain.MemberTypeNormal, domain.LeaderTypeMember,
		)
	}
	return members
}

// BenchmarkSaveBatch_LoopVsBulk 는 변경 전(개별 Save 루프)과 변경 후(CreateInBatches)의
// 성능을 멤버 수별로 비교합니다.
//
// 실행:
//
//	go test -bench=BenchmarkSaveBatch_LoopVsBulk -benchmem -count=3 -timeout=10m \
//	  ./test/contest/infra/persistence/...
func BenchmarkSaveBatch_LoopVsBulk(b *testing.B) {
	ctx := context.Background()
	container, err := support.SetupMySQLContainer(ctx)
	if err != nil {
		b.Fatal(err)
	}
	defer container.Teardown(ctx)

	db := container.GetDB()
	if err := db.AutoMigrate(&domain.Contest{}, &domain.ContestMember{}); err != nil {
		b.Fatal(err)
	}

	contestAdapter := adapter.NewContestDatabaseAdapter(db)

	sizes := []int{10, 50, 100, 200}

	for _, size := range sizes {
		// 변경 전: 개별 Save 루프
		b.Run(fmt.Sprintf("LoopSave/%d_members", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				contest := createBenchContest(contestAdapter, fmt.Sprintf("loop-%d-%d", size, i))
				members := makeMembers(size, contest.ContestID, i*10000)
				b.StartTimer()

				saveBatchLoop(db, members)
			}
		})

		// 변경 후: CreateInBatches
		b.Run(fmt.Sprintf("CreateInBatches/%d_members", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				contest := createBenchContest(contestAdapter, fmt.Sprintf("bulk-%d-%d", size, i))
				members := makeMembers(size, contest.ContestID, i*10000)
				b.StartTimer()

				saveBatchBulk(db, members)
			}
		})
	}
}
