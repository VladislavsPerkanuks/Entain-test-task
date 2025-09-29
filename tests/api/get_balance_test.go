package api

type GetBalanceTestSuite struct {
	APITestSuite
}

// TestGetBalanceUser1 tests the balance retrieval for user 1 => should be 100.00.
func (s *GetBalanceTestSuite) TestGetBalanceUser1() {
	resp := s.GetBalance(s.T(), 1)

	expected := `
{
	"userId": 1,
	"balance": "100.00"
}`

	s.Equal(200, resp.StatusCode, "Should return 200 OK")
	s.JSONEq(expected, string(resp.Body), "Response body should match expected JSON")
}

// TestGetBalanceUser2 tests the balance retrieval for user 2 => should be 200.00.
func (s *GetBalanceTestSuite) TestGetBalanceUser2() {
	resp := s.GetBalance(s.T(), 2)

	expected := `
{
	"userId": 2,
	"balance": "200.00"
}`

	s.Equal(200, resp.StatusCode, "Should return 200 OK")
	s.JSONEq(expected, string(resp.Body), "Response body should match expected JSON")
}

// TestGetBalanceUser3 tests the balance retrieval for user 3 => should be 50.00.
func (s *GetBalanceTestSuite) TestGetBalanceUser3() {
	resp := s.GetBalance(s.T(), 3)

	expected := `
{
	"userId": 3,
	"balance": "50.00"
}`

	s.Equal(200, resp.StatusCode, "Should return 200 OK")
	s.JSONEq(expected, string(resp.Body), "Response body should match expected JSON")
}

// TestGetBalanceUser4 tests the balance retrieval for user 4 => should be 33.33.
func (s *GetBalanceTestSuite) TestGetBalanceUser4() {
	resp := s.GetBalance(s.T(), 4)

	expected := `
{
	"userId": 4,
	"balance": "33.33"
}`

	s.Equal(200, resp.StatusCode, "Should return 200 OK")
	s.JSONEq(expected, string(resp.Body), "Response body should match expected JSON")
}

// TestGetBalanceNonExistentUser tests the balance retrieval for a non-existent user.
func (s *GetBalanceTestSuite) TestGetBalanceNonExistentUser() {
	resp := s.GetBalance(s.T(), 999)

	s.NotEqual(200, resp.StatusCode, "Should return error status for non-existent user")
}
