package api

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

type PerformanceTestSuite struct {
	APITestSuite
}

type LoadTestConfig struct {
	TargetRPS int
	Duration  time.Duration
}

type LoadTestResult struct {
	TotalRequests   int
	SuccessCount    int32
	ErrorCount      int32
	ActualRPS       float64
	SuccessRate     float64
	TotalDuration   time.Duration
	AvgResponseTime time.Duration
	MinResponseTime time.Duration
	MaxResponseTime time.Duration
	ResponseTimes   []time.Duration
}

// TestLoadPerformance50RPS tests if the app can handle 50 requests per second
func (s *PerformanceTestSuite) TestLoadPerformance50RPS() {
	s.T().Log("Starting load test: 50 requests per second for 10 seconds")

	config := LoadTestConfig{
		TargetRPS: 50,
		Duration:  10 * time.Second,
	}

	result := s.executeLoadTest(config)
	s.logTestResults(config, result)
	s.assertPerformanceRequirements(config, result)
}

func (s *PerformanceTestSuite) executeLoadTest(config LoadTestConfig) LoadTestResult {
	totalRequests := int(config.TargetRPS * int(config.Duration.Seconds()))

	// Results tracking
	var successCount, errorCount int32
	var responseTimes []time.Duration
	var mu sync.Mutex

	ticker := time.NewTicker(time.Second / time.Duration(config.TargetRPS))
	defer ticker.Stop()

	var wg sync.WaitGroup
	startTime := time.Now()

	s.T().Logf("Sending %d requests at %d RPS over %v", totalRequests, config.TargetRPS, config.Duration)

	// Send requests at controlled rate
	for i := range totalRequests {
		select {
		case <-ticker.C:
			wg.Add(1)
			go s.executeRequest(&wg, i, &successCount, &errorCount, &responseTimes, &mu)
		case <-time.After(config.Duration + time.Second):
			s.T().Log("Load test timeout reached")
			goto waitForCompletion
		}
	}

waitForCompletion:
	s.waitForRequestsCompletion(&wg)

	totalTime := time.Since(startTime)
	actualRequests := len(responseTimes)
	actualRPS := float64(actualRequests) / totalTime.Seconds()

	stats := s.calculateResponseTimeStats(responseTimes)

	return LoadTestResult{
		TotalRequests:   actualRequests,
		SuccessCount:    successCount,
		ErrorCount:      errorCount,
		ActualRPS:       actualRPS,
		SuccessRate:     float64(successCount) / float64(actualRequests) * 100,
		TotalDuration:   totalTime,
		AvgResponseTime: stats.avg,
		MinResponseTime: stats.min,
		MaxResponseTime: stats.max,
		ResponseTimes:   responseTimes,
	}
}

func (s *PerformanceTestSuite) executeRequest(
	wg *sync.WaitGroup,
	requestNum int,
	successCount, errorCount *int32,
	responseTimes *[]time.Duration,
	mu *sync.Mutex,
) {
	defer wg.Done()

	reqStart := time.Now()
	success := s.performRequestOperation(requestNum)
	requestDuration := time.Since(reqStart)

	mu.Lock()
	*responseTimes = append(*responseTimes, requestDuration)
	if success {
		atomic.AddInt32(successCount, 1)
	} else {
		atomic.AddInt32(errorCount, 1)
	}

	mu.Unlock()
}

func (s *PerformanceTestSuite) performRequestOperation(requestNum int) bool {
	userID := (requestNum % 4) + 1 // Users 1, 2, 3, 4

	if requestNum%4 == 0 {
		// 25% balance checks
		return s.performBalanceCheck(userID)
	}

	// 75% transactions (mix of win/lose)
	return s.performTransaction(userID, requestNum)
}

func (s *PerformanceTestSuite) performBalanceCheck(userID int) bool {
	resp := s.GetBalance(s.T(), userID)

	return resp.StatusCode == 200
}

func (s *PerformanceTestSuite) performTransaction(userID, requestNum int) bool {
	state := "lose"
	amount := "0.13"

	if requestNum%2 == 0 {
		state = "win"
		amount = "0.49"
	}

	transactionReq := TransactionRequest{
		State:         state,
		Amount:        amount,
		TransactionID: uuid.New().String(),
	}

	sourceTypes := []string{"game", "server", "payment"}
	sourceType := sourceTypes[requestNum%3]

	resp := s.ProcessTransaction(s.T(), userID, sourceType, transactionReq)

	return resp.StatusCode == 200
}

func (s *PerformanceTestSuite) waitForRequestsCompletion(wg *sync.WaitGroup) {
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		s.T().Log("All requests completed")
	case <-time.After(30 * time.Second):
		s.T().Log("Timeout waiting for requests to complete")
	}
}

type responseTimeStats struct {
	avg, min, max time.Duration
}

func (s *PerformanceTestSuite) calculateResponseTimeStats(responseTimes []time.Duration) responseTimeStats {
	if len(responseTimes) == 0 {
		return responseTimeStats{}
	}

	var totalDuration time.Duration
	minDuration := time.Hour
	var maxDuration time.Duration

	for _, duration := range responseTimes {
		totalDuration += duration

		if duration > maxDuration {
			maxDuration = duration
		}

		if duration < minDuration {
			minDuration = duration
		}
	}

	avgDuration := totalDuration / time.Duration(len(responseTimes))

	return responseTimeStats{
		avg: avgDuration,
		min: minDuration,
		max: maxDuration,
	}
}

func (s *PerformanceTestSuite) logTestResults(config LoadTestConfig, result LoadTestResult) {
	s.T().Logf("=== LOAD TEST RESULTS ===")
	s.T().Logf("Target RPS: %d", config.TargetRPS)
	s.T().Logf("Actual RPS: %.2f", result.ActualRPS)
	s.T().Logf("Total Requests: %d", result.TotalRequests)
	s.T().Logf("Successful Requests: %d", result.SuccessCount)
	s.T().Logf("Failed Requests: %d", result.ErrorCount)
	s.T().Logf("Success Rate: %.2f%%", result.SuccessRate)
	s.T().Logf("Total Test Duration: %v", result.TotalDuration)
	s.T().Logf("Average Response Time: %v", result.AvgResponseTime)
	s.T().Logf("Min Response Time: %v", result.MinResponseTime)
	s.T().Logf("Max Response Time: %v", result.MaxResponseTime)
}

func (s *PerformanceTestSuite) assertPerformanceRequirements(config LoadTestConfig, result LoadTestResult) {
	minAcceptableRPS := float64(config.TargetRPS) * 0.8
	s.True(
		result.ActualRPS >= minAcceptableRPS,
		"Should achieve at least 80%% of target RPS (%.2f >= %.2f)",
		result.ActualRPS, minAcceptableRPS,
	)

	s.True(
		result.SuccessRate >= 95.0,
		"Should have at least 95%% success rate (%.2f%% >= 95%%)",
		result.SuccessRate,
	)

	s.True(
		result.AvgResponseTime < 1*time.Second,
		"Average response time should be under 1 second (%v < 1s)",
		result.AvgResponseTime,
	)

	s.True(
		result.MaxResponseTime < 5*time.Second,
		"Max response time should be under 5 seconds (%v < 5s)",
		result.MaxResponseTime,
	)
}
