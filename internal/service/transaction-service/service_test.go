package transaction_service

import (
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"
	"transaction-monitoring-system/internal/dto"
	"transaction-monitoring-system/internal/repository"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockRepository реализует интерфейс репозитория, используемый TransactionService.
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) GetActiveFraudRules() ([]dto.FraudRuleDTO, error) {
	args := m.Called()
	return args.Get(0).([]dto.FraudRuleDTO), args.Error(1)
}

func (m *MockRepository) CreateTransaction(transaction dto.TransactionDTO) (int64, error) {
	args := m.Called(transaction)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRepository) CreateDoubtfulTransaction(doubtful dto.DoubtfulTransactionDTO) error {
	args := m.Called(doubtful)
	return args.Error(0)
}

func (m *MockRepository) GetTransactionById(id int64) (dto.TransactionDTO, error) {
	return dto.TransactionDTO{}, nil
}
func (m *MockRepository) GetTransactionsByKey(key string) ([]dto.TransactionDTO, error) {
	return []dto.TransactionDTO{}, nil
}
func (m *MockRepository) UpdateTransactionStatusById(transactionId int64, status string) error {
	return nil
}
func (m *MockRepository) UpdateDecisionByTrId(transactionId int64, decision string) error {
	return nil
}
func (m *MockRepository) GetDoubtfulTransactionsByKey(key string) ([]dto.DoubtfulTransactionDTO, error) {
	return []dto.DoubtfulTransactionDTO{}, nil
}
func (m *MockRepository) DeleteDoubtfulTransactionByDecision(decision string) error {
	return nil
}
func (m *MockRepository) GetFraudRulesByKey(key string) ([]dto.FraudRuleDTO, error) {
	return []dto.FraudRuleDTO{}, nil
}
func (m *MockRepository) UpdateFraudRule(rule dto.FraudRuleDTO) error {
	return nil
}
func (m *MockRepository) CreateFraudRule(rule dto.FraudRuleDTO) error {
	return nil
}
func (m *MockRepository) DeleteFraudRuleById(ruleId int64) error {
	return nil
}

func TestTransactionService_Control(t *testing.T) {
	cases := []struct {
		name      string
		prepare   func(m *MockRepository)
		tx        dto.TransactionDTO
		respError string
		check     func(t *testing.T, m *MockRepository)
	}{
		{
			name: "Success innocent decision",
			prepare: func(m *MockRepository) {
				rules := []dto.FraudRuleDTO{
					{Name: "high_amount", FieldName: "amount", Operator: ">", Value: "100", AddRisk: 30},
				}
				m.On("GetActiveFraudRules").Return(rules, nil)
				m.On("CreateTransaction", mock.MatchedBy(func(tx dto.TransactionDTO) bool {
					return tx.Status == Approved
				})).Return(int64(1), nil)
			},
			tx: dto.TransactionDTO{
				Hash:      "innocent_hash",
				Source:    "clean",
				Amount:    50,
				Direction: "in",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			respError: "",
			check: func(t *testing.T, m *MockRepository) {
				m.AssertNotCalled(t, "CreateDoubtfulTransaction")
			},
		},
		{
			name: "Success review decision",
			prepare: func(m *MockRepository) {
				rules := []dto.FraudRuleDTO{
					{Name: "high_amount", FieldName: "amount", Operator: ">", Value: "100", AddRisk: 30},
				}
				m.On("GetActiveFraudRules").Return(rules, nil)
				m.On("CreateTransaction", mock.MatchedBy(func(tx dto.TransactionDTO) bool {
					return tx.Status == Pending
				})).Return(int64(2), nil)
				m.On("CreateDoubtfulTransaction", mock.MatchedBy(func(d dto.DoubtfulTransactionDTO) bool {
					return d.Decision == Review && d.TransactionId == 2
				})).Return(nil)
			},
			tx: dto.TransactionDTO{
				Hash:      "review_hash",
				Source:    "good",
				Amount:    150,
				Direction: "out",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			respError: "",
			check: func(t *testing.T, m *MockRepository) {
				m.AssertCalled(t, "CreateDoubtfulTransaction", mock.Anything)
			},
		},
		{
			name: "Success block decision",
			prepare: func(m *MockRepository) {
				rules := []dto.FraudRuleDTO{
					{Name: "high_amount", FieldName: "amount", Operator: ">", Value: "100", AddRisk: 30},
					{Name: "bad_source", FieldName: "source", Operator: "=", Value: "bad", AddRisk: 60},
				}
				m.On("GetActiveFraudRules").Return(rules, nil)
				m.On("CreateTransaction", mock.MatchedBy(func(tx dto.TransactionDTO) bool {
					return tx.Status == Rejected
				})).Return(int64(3), nil)
				m.On("CreateDoubtfulTransaction", mock.MatchedBy(func(d dto.DoubtfulTransactionDTO) bool {
					return d.Decision == Block && d.TransactionId == 3
				})).Return(nil)
			},
			tx: dto.TransactionDTO{
				Hash:      "block_hash",
				Source:    "bad",
				Amount:    200,
				Direction: "in",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			respError: "",
			check: func(t *testing.T, m *MockRepository) {
				m.AssertCalled(t, "CreateDoubtfulTransaction", mock.Anything)
			},
		},
		{
			name: "Error no active fraud rules",
			prepare: func(m *MockRepository) {
				m.On("GetActiveFraudRules").Return([]dto.FraudRuleDTO{}, errors.New("internal.repository.postgres.fraud-rule.GetActiveFrudRules : "+repository.ErrRecordNotFound.Error()))
			},
			tx: dto.TransactionDTO{
				Hash:      "no_rules_hash",
				Source:    "any",
				Amount:    10,
				Direction: "in",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			respError: "internal.repository.postgres.fraud-rule.GetActiveFrudRules : " + repository.ErrRecordNotFound.Error(),
			check: func(t *testing.T, m *MockRepository) {
				m.AssertNotCalled(t, "CreateTransaction")
				m.AssertNotCalled(t, "CreateDoubtfulTransaction")
			},
		},
		{
			name: "Error duplicate transaction hash",
			prepare: func(m *MockRepository) {
				rules := []dto.FraudRuleDTO{
					{Name: "any", FieldName: "amount", Operator: ">", Value: "0", AddRisk: 0},
				}
				m.On("GetActiveFraudRules").Return(rules, nil)
				m.On("CreateTransaction", mock.Anything).Return(int64(0), errors.New("internal.repository.postgres.transaction.CreateTransaction : "+repository.ErrRecordAlreadyExists.Error()))
			},
			tx: dto.TransactionDTO{
				Hash:      "dup_hash",
				Source:    "src",
				Amount:    10,
				Direction: "dir",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			respError: "internal.repository.postgres.transaction.CreateTransaction : " + repository.ErrRecordAlreadyExists.Error(),
			check: func(t *testing.T, m *MockRepository) {
				m.AssertNotCalled(t, "CreateDoubtfulTransaction")
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			svc := NewTransactionService(logger, mockRepo)

			tc.prepare(mockRepo)

			err := svc.Control(tc.tx)

			if tc.respError == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.respError)
			}

			if tc.check != nil {
				tc.check(t, mockRepo)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func Test_inOutList(t *testing.T) {
	cases := []struct {
		name      string
		target    string
		operator  string
		listValue string
		want      bool
		wantErr   bool
	}{
		{
			name:      "in: target present",
			target:    "a",
			operator:  "in",
			listValue: "a,b,c",
			want:      true,
		},
		{
			name:      "in: target present with spaces",
			target:    "a",
			operator:  "in",
			listValue: " b , a , c ",
			want:      true,
		},
		{
			name:      "in: target not present",
			target:    "z",
			operator:  "in",
			listValue: "a,b,c",
			want:      false,
		},
		{
			name:      "out: target not present -> true",
			target:    "z",
			operator:  "out",
			listValue: "a,b,c",
			want:      true,
		},
		{
			name:      "out: target present -> false",
			target:    "b",
			operator:  "out",
			listValue: "a,b,c",
			want:      false,
		},
		{
			name:      "empty list, in: not present",
			target:    "x",
			operator:  "in",
			listValue: "",
			want:      false,
		},
		{
			name:      "empty list, out: not present -> true",
			target:    "x",
			operator:  "out",
			listValue: "",
			want:      true,
		},
		{
			name:      "invalid operator",
			target:    "a",
			operator:  "unknown",
			listValue: "a,b",
			wantErr:   true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, err := inOutList(tc.target, tc.operator, tc.listValue)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.want, got)
			}
		})
	}
}

func Test_like(t *testing.T) {
	cases := []struct {
		name   string
		str    string
		subStr string
		want   bool
	}{
		{
			name:   "substring present",
			str:    "hello world",
			subStr: "world",
			want:   true,
		},
		{
			name:   "substring not present",
			str:    "hello",
			subStr: "bye",
			want:   false,
		},
		{
			name:   "empty substring",
			str:    "abc",
			subStr: "",
			want:   true,
		},
		{
			name:   "empty string, non-empty substring",
			str:    "",
			subStr: "a",
			want:   false,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, err := like(tc.str, tc.subStr)
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}
