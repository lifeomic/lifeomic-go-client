package client

type graphqlClient interface {
	Gql(string, string, map[string]interface{}) (*map[string]interface{}, error)
}
