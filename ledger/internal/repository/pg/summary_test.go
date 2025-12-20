package pg

import (
	"context"
	"database/sql"
	"testing"

	"ledger/internal/domain"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSummaryPgRepository_GetSummary(t *testing.T) {
	db, dbMock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer redisClient.Close()

	repo := NewSummaryPgRepository(db, redisClient)
	ctx := context.Background()

	tests := []struct {
		name        string
		from        string
		to          string
		mockSetup   func()
		expected    *domain.Summary
		expectedErr bool
	}{
		{
			name: "cache hit",
			from: "2025-12-01",
			to:   "2025-12-31",
			mockSetup: func() {
				mr.Set("report:summary:2025-12-01:2025-12-31", `{"Food":500,"Transport":200}`)
			},
			expected: &domain.Summary{
				Categories:  map[string]float64{"Food": 500.0, "Transport": 200.0},
				CacheResult: true,
			},
			expectedErr: false,
		},
		{
			name: "database error on categories query",
			from: "2025-12-01",
			to:   "2025-12-31",
			mockSetup: func() {
				dbMock.ExpectQuery(`SELECT distinct category FROM expenses WHERE date BETWEEN \$1 AND \$2`).
					WithArgs("2025-12-01", "2025-12-31").
					WillReturnError(sql.ErrConnDone)
			},
			expected:    nil,
			expectedErr: true,
		},
		{
			name: "no categories found",
			from: "2025-12-01",
			to:   "2025-12-31",
			mockSetup: func() {
				categoryRows := sqlmock.NewRows([]string{"category"})
				dbMock.ExpectQuery(`SELECT distinct category FROM expenses WHERE date BETWEEN \$1 AND \$2`).
					WithArgs("2025-12-01", "2025-12-31").
					WillReturnRows(categoryRows)
			},
			expected: &domain.Summary{
				Categories:  map[string]float64{},
				CacheResult: false,
			},
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr.FlushAll()

			tt.mockSetup()

			result, err := repo.GetSummary(ctx, tt.from, tt.to)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}

			if tt.name != "successful database query (cache miss)" {
				assert.NoError(t, dbMock.ExpectationsWereMet())
			}
		})
	}
}
