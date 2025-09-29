package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/VladislavsPerkanuks/Entain-test-task/internal/config"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// APITestSuite provides test setup and teardown for All API integration tests
type APITestSuite struct {
	suite.Suite
	httpClient *http.Client
	testDB     *sql.DB
	BaseURL    string
	testConfig *config.Config
}

// apiResponse represents a minimal HTTP response wrapper used in tests.
type apiResponse struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

// TransactionRequest represents the payload sent to the transaction endpoint in tests.
type TransactionRequest struct {
	State         string `json:"state"`
	Amount        string `json:"amount"`
	TransactionID string `json:"transactionId"`
}

// BalanceResponse mirrors the public contract of the balance endpoint.
type BalanceResponse struct {
	UserID  int    `json:"userId"`
	Balance string `json:"balance"`
}

// SetupSuite runs once before all tests in the suite
func (s *APITestSuite) SetupSuite() {
	s.testConfig = config.DefaultConfig()
	s.BaseURL = "http://localhost:3000"

	s.httpClient = &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        200,
			MaxIdleConnsPerHost: 200,
			MaxConnsPerHost:     200,
		},
	}

	s.Require().NoError(s.waitForServer(s.BaseURL, 30*time.Second), "server not ready")

	connectionString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", s.testConfig.DB_USER, s.testConfig.DB_PASSWORD, s.testConfig.DB_HOST, s.testConfig.DB_PORT, s.testConfig.DB_NAME)

	conn, err := sql.Open("postgres", connectionString)
	s.Require().NoError(err, "failed to open test database connection")
	s.testDB = conn
}

// SetupTest runs before each individual test
func (s *APITestSuite) SetupTest() {
	s.resetDatabaseState(context.Background())
}

func TestTransactionTestSuite(t *testing.T) {
	suite.Run(t, new(TransactionTestSuite))
}

func TestGetBalanceTestSuite(t *testing.T) {
	suite.Run(t, new(GetBalanceTestSuite))
}

func TestPerformanceTestSuite(t *testing.T) {
	if _, ok := os.LookupEnv("RUN_PERFORMANCE_TESTS"); !ok {
		t.Skip("Skipping performance tests. Set RUN_PERFORMANCE_TESTS=1 to run.")
	}

	suite.Run(t, new(PerformanceTestSuite))
}

// GetBalance calls the GET /user/{id}/balance endpoint and returns the parsed payload when successful.
func (s *APITestSuite) GetBalance(tb testing.TB, userID int) apiResponse {
	tb.Helper()

	url := fmt.Sprintf("%s/user/%d/balance", strings.TrimRight(s.BaseURL, "/"), userID)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(tb, err, "Failed to create GET request for user %d", userID)

	return s.performRequest(req)
}

// ProcessTransaction performs a POST /user/{id}/transaction request with the provided payload and source type.
func (s *APITestSuite) ProcessTransaction(tb testing.TB, userID int, sourceType string, body TransactionRequest) apiResponse {
	tb.Helper()

	url := fmt.Sprintf("%s/user/%d/transaction", strings.TrimRight(s.BaseURL, "/"), userID)

	payload, err := json.Marshal(body)
	s.Require().NoError(err, "failed to marshal transaction request body")

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(payload))
	s.Require().NoError(err, "failed to create POST request for user %d transaction", userID)

	req.Header.Set("Content-Type", "application/json")
	if sourceType != "" {
		req.Header.Set("Source-Type", sourceType)
	}

	return s.performRequest(req)
}

func (s *APITestSuite) performRequest(req *http.Request) apiResponse {
	resp, err := s.httpClient.Do(req)
	s.Require().NoError(err)

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)

	return apiResponse{StatusCode: resp.StatusCode, Headers: resp.Header.Clone(), Body: body}
}

func (s *APITestSuite) resetDatabaseState(ctx context.Context) {
	if ctx == nil {
		ctx = context.Background()
	}

	statements := []string{
		"TRUNCATE TABLE transactions RESTART IDENTITY CASCADE",
		"TRUNCATE TABLE users RESTART IDENTITY CASCADE",
		"INSERT INTO users (balance) VALUES (100.00), (200.00), (50.00), (33.33)",
	}

	tx, err := s.testDB.BeginTx(ctx, nil)
	s.Require().NoError(err, "failed to begin transaction for database reset")

	defer func() {
		_ = tx.Rollback()
	}()

	for _, stmt := range statements {
		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			s.Require().NoError(err, "failed to execute statement")
		}
	}

	s.Require().NoError(tx.Commit(), "failed to commit reset transaction")
}

func (s *APITestSuite) waitForServer(baseURL string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for {
		if time.Now().After(deadline) {
			return fmt.Errorf("server did not become ready in %s", timeout)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/user/1/balance", strings.TrimRight(baseURL, "/")), nil)
		if err != nil {
			cancel()
			return fmt.Errorf("create readiness request: %w", err)
		}

		resp, err := s.httpClient.Do(req)
		cancel()
		if err == nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()

			if resp.StatusCode != http.StatusNotFound {
				return nil
			}
		}

		time.Sleep(500 * time.Millisecond)
	}
}
