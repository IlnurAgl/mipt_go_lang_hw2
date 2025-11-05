package ledger

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTransactions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		amount   float64
		category string
		err      error
	}{
		{name: "0", amount: 0, category: "", err: errors.New("invalid amount")},
		{name: "valid", amount: 10, category: "еда", err: nil},
		{name: "negative", amount: -10, category: "еда", err: errors.New("invalid amount")},
		{name: "invalid_category", amount: 10, category: "", err: errors.New("invalid category")},
	}
	ReadBudget()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tr := Transaction{
				ID:          GetTransactionId(),
				Category:    tc.category,
				Amount:      tc.amount,
				Description: "test",
				Date:        time.Now().String(),
			}
			err := tr.Validate()
			if tc.err != nil {
				assert.EqualError(t, tc.err, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
	t.Cleanup(func() { Reset() })
}

func TestBudgets(t *testing.T) {
	err := SetBudget(Budget{Category: "еда", Limit: 5000})
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	err = AddTransaction(Transaction{
		ID:          GetTransactionId(),
		Category:    "еда",
		Amount:      5000,
		Description: "test",
		Date:        time.Now().String(),
	})
	assert.NoError(t, err)
	assert.Equal(t, len(ListTransactions()), 1)
	err = AddTransaction(Transaction{
		ID:          GetTransactionId(),
		Category:    "еда",
		Amount:      5001,
		Description: "test",
		Date:        time.Now().String(),
	})
	assert.EqualError(t, err, errors.New("budget exceeded").Error())
	assert.Equal(t, len(ListTransactions()), 1)
}
