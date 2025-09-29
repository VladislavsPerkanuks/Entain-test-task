package api

import (
	"github.com/google/uuid"
)

type TransactionTestSuite struct {
	APITestSuite
}

// TestProcessTransactionWin tests win transactions that increase user balance.
func (s *TransactionTestSuite) TestProcessTransactionWinGame() {
	transactionReq := TransactionRequest{
		State:         "win",
		Amount:        "10.15",
		TransactionID: uuid.New().String(),
	}

	resp := s.ProcessTransaction(s.T(), 1, "game", transactionReq)
	s.Equal(200, resp.StatusCode, "Win transaction should return 200 OK")

	// Check balance was updated correctly (100.00 + 10.15 = 110.15)
	balanceResp := s.GetBalance(s.T(), 1)
	s.Equal(200, balanceResp.StatusCode, "Balance request should return 200 OK")

	expected := `{"userId": 1, "balance": "110.15"}`
	s.JSONEq(expected, string(balanceResp.Body), "Balance should be increased by win amount")
}

func (s *TransactionTestSuite) TestProcessTransactionWinServer() {
	transactionReq := TransactionRequest{
		State:         "win",
		Amount:        "25.50",
		TransactionID: uuid.New().String(),
	}

	resp := s.ProcessTransaction(s.T(), 2, "server", transactionReq)
	s.Equal(200, resp.StatusCode, "Win transaction should return 200 OK")

	// Check balance was updated correctly (200.00 + 25.50 = 225.50)
	balanceResp := s.GetBalance(s.T(), 2)
	expected := `{"userId": 2, "balance": "225.50"}`
	s.JSONEq(expected, string(balanceResp.Body), "Balance should be increased by win amount")
}

func (s *TransactionTestSuite) TestProcessTransactionWinPayment() {
	transactionReq := TransactionRequest{
		State:         "win",
		Amount:        "5.99",
		TransactionID: uuid.New().String(),
	}

	resp := s.ProcessTransaction(s.T(), 3, "payment", transactionReq)
	s.Equal(200, resp.StatusCode, "Win transaction should return 200 OK")

	// Check balance was updated correctly (50.00 + 5.99 = 55.99)
	balanceResp := s.GetBalance(s.T(), 3)
	expected := `{"userId": 3, "balance": "55.99"}`
	s.JSONEq(expected, string(balanceResp.Body), "Balance should be increased by win amount")
}

// Test lose transaction that decreases user balance.
func (s *TransactionTestSuite) TestProcessTransactionLose() {
	transactionReq := TransactionRequest{
		State:         "lose",
		Amount:        "15.25",
		TransactionID: uuid.New().String(),
	}

	resp := s.ProcessTransaction(s.T(), 1, "game", transactionReq)
	s.Equal(200, resp.StatusCode, "Lose transaction should return 200 OK")

	// Check balance was updated correctly (100.00 - 15.25 = 84.75)
	balanceResp := s.GetBalance(s.T(), 1)
	expected := `{"userId": 1, "balance": "84.75"}`
	s.JSONEq(expected, string(balanceResp.Body), "Balance should be decreased by lose amount")
}

// Test lose transaction that would result in negative balance.
func (s *TransactionTestSuite) TestProcessTransactionLoseInsufficientBalance() {
	transactionReq := TransactionRequest{
		State:         "lose",
		Amount:        "150.00", // More than user 1's balance of 100.00
		TransactionID: "test-lose-insufficient-001",
	}

	resp := s.ProcessTransaction(s.T(), 1, "game", transactionReq)
	s.NotEqual(200, resp.StatusCode, "Should return error for insufficient balance")

	// Balance should remain unchanged
	balanceResp := s.GetBalance(s.T(), 1)
	expected := `{"userId": 1, "balance": "100.00"}`
	s.JSONEq(expected, string(balanceResp.Body), "Balance should remain unchanged on failed transaction")
}

func (s *TransactionTestSuite) TestProcessTransactionDuplicateTransactionId() {
	transactionReq := TransactionRequest{
		State:         "win",
		Amount:        "10.00",
		TransactionID: uuid.New().String(),
	}

	// Process first transaction
	resp1 := s.ProcessTransaction(s.T(), 1, "game", transactionReq)
	s.Equal(200, resp1.StatusCode, "First transaction should succeed")

	// Check balance after first transaction (100.00 + 10.00 = 110.00)
	balanceResp := s.GetBalance(s.T(), 1)
	expected := `{"userId": 1, "balance": "110.00"}`
	s.JSONEq(expected, string(balanceResp.Body), "Balance should be updated after first transaction")

	// Process same transaction again (should be ignored)
	s.ProcessTransaction(s.T(), 1, "game", transactionReq)

	// Check balance remains the same (shouldn't double-process)
	balanceResp2 := s.GetBalance(s.T(), 1)
	s.JSONEq(expected, string(balanceResp2.Body), "Balance should not change on duplicate transaction")
}

func (s *TransactionTestSuite) TestProcessTransactionNonExistingUserId() {
	transactionReq := TransactionRequest{
		State:         "win",
		Amount:        "10.00",
		TransactionID: uuid.New().String(),
	}

	resp := s.ProcessTransaction(s.T(), 999, "game", transactionReq)
	s.NotEqual(200, resp.StatusCode, "Should return error for non-existent user")
}

func (s *TransactionTestSuite) TestProcessTransactionMissingSourceType() {
	transactionReq := TransactionRequest{
		State:         "win",
		Amount:        "10.00",
		TransactionID: uuid.New().String(),
	}

	// Pass empty source type
	resp := s.ProcessTransaction(s.T(), 1, "", transactionReq)
	s.NotEqual(200, resp.StatusCode, "Should return error for missing Source-Type header")
}

func (s *TransactionTestSuite) TestProcessTransactionInvalidState() {
	transactionReq := TransactionRequest{
		State:         "invalid",
		Amount:        "10.00",
		TransactionID: uuid.New().String(),
	}

	resp := s.ProcessTransaction(s.T(), 1, "game", transactionReq)
	s.NotEqual(200, resp.StatusCode, "Should return error for invalid state")
}

func (s *TransactionTestSuite) TestProcessTransactionInvalidAmount() {
	// Test with negative amount
	transactionReq := TransactionRequest{
		State:         "win",
		Amount:        "-10.00",
		TransactionID: uuid.New().String(),
	}

	resp := s.ProcessTransaction(s.T(), 1, "game", transactionReq)
	s.NotEqual(200, resp.StatusCode, "Should return error for negative amount")
}

func (s *TransactionTestSuite) TestProcessTransactionInvalidAmountFormat() {
	// Test with invalid amount format
	transactionReq := TransactionRequest{
		State:         "win",
		Amount:        "abc",
		TransactionID: uuid.New().String(),
	}

	resp := s.ProcessTransaction(s.T(), 1, "game", transactionReq)
	s.NotEqual(200, resp.StatusCode, "Should return error for invalid amount format")
}

func (s *TransactionTestSuite) TestProcessTransactionWorkflow() {
	// Complete workflow test: multiple transactions and balance checks

	// Initial balance for user 1 should be 100.00
	balanceResp := s.GetBalance(s.T(), 1)
	s.Equal(200, balanceResp.StatusCode)
	expected := `{"userId": 1, "balance": "100.00"}`
	s.JSONEq(expected, string(balanceResp.Body))

	// Win 25.50
	winReq := TransactionRequest{
		State:         "win",
		Amount:        "25.50",
		TransactionID: uuid.New().String(),
	}
	resp := s.ProcessTransaction(s.T(), 1, "game", winReq)
	s.Equal(200, resp.StatusCode)

	// Check balance: 100.00 + 25.50 = 125.50
	balanceResp = s.GetBalance(s.T(), 1)
	expected = `{"userId": 1, "balance": "125.50"}`
	s.JSONEq(expected, string(balanceResp.Body))

	// Lose 15.25
	loseReq := TransactionRequest{
		State:         "lose",
		Amount:        "15.25",
		TransactionID: uuid.New().String(),
	}
	resp = s.ProcessTransaction(s.T(), 1, "server", loseReq)
	s.Equal(200, resp.StatusCode)

	// Check final balance: 125.50 - 15.25 = 110.25
	balanceResp = s.GetBalance(s.T(), 1)
	expected = `{"userId": 1, "balance": "110.25"}`
	s.JSONEq(expected, string(balanceResp.Body))
}

func (s *TransactionTestSuite) TestProcessTransactionDecimalPrecision() {
	// Test with amounts having exactly 2 decimal places
	transactionReq := TransactionRequest{
		State:         "win",
		Amount:        "0.01", // Minimum amount
		TransactionID: uuid.New().String(),
	}

	resp := s.ProcessTransaction(s.T(), 1, "game", transactionReq)
	s.Equal(200, resp.StatusCode, "Should handle small decimal amounts")

	// Check balance: 100.00 + 0.01 = 100.01
	balanceResp := s.GetBalance(s.T(), 1)
	expected := `{"userId": 1, "balance": "100.01"}`
	s.JSONEq(expected, string(balanceResp.Body))
}
