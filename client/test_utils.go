package client

type MockClient struct {
	hasBeenCalled bool
	response      *map[string]interface{}
	error         error
}

func (m *MockClient) Gql(url string, operation string, variables map[string]interface{}) (*map[string]interface{}, error) {
	m.hasBeenCalled = true
	return m.response, m.error
}
