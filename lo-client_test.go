package lo_client

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
)

const MOCK_MUTATION = `
mutation MockMutation($var: String!) {
	some_mutation(var: $var) {
		result
	}
}
`

func TestBuildGqlQuery(t *testing.T) {
	client := Client{
		rules: map[string]bool{
			"testRule": true,
		},
	}
	raw := client.buildGqlQuery("/some/path", MOCK_MUTATION, map[string]interface{}{"var": "value"})
	var parsed map[string]interface{}
	err := json.Unmarshal(raw, &parsed)
	if err != nil {
		t.Fatal("Could not parse payload as json", string(raw))
	}
	var parsedBody map[string]interface{}
	err = json.Unmarshal([]byte(parsed["body"].(string)), &parsedBody)
	if err != nil {
		t.Fatal("Could not parse body as json", parsed["body"])
	}
	if parsedBody["query"].(string) != MOCK_MUTATION {
		t.Fatal("Missing query in body", parsed["body"])
	}
	variables := parsedBody["variables"].(map[string]interface{})
	if variables["var"] != "value" {
		t.Fatal("Missing variable", variables)
	}

	headers := parsed["headers"].(map[string]interface{})

	var lifeomicPolicy struct {
		Rules map[string]bool
	}
	err = json.Unmarshal([]byte(headers["LifeOmic-Policy"].(string)), &lifeomicPolicy)
	if !lifeomicPolicy.Rules["testRule"] {
		t.Fatal("Missing rule testRule in rules", lifeomicPolicy)
	}

	path := parsed["path"].(string)
	if path != "/some/path" {
		t.Fatal("Did not use correct path", path)
	}
}

type MockInvoker struct {
	hasBeenCalled bool
	payload       *lambda.InvokeInput
	response      *lambda.InvokeOutput
	err           error
}

func (m *MockInvoker) Invoke(ctx context.Context, payload *lambda.InvokeInput, rest ...func(*lambda.Options)) (*lambda.InvokeOutput, error) {
	m.hasBeenCalled = true
	m.payload = payload
	return m.response, m.err
}

func TestGql(t *testing.T) {
	mock := MockInvoker{
		response: &lambda.InvokeOutput{
			Payload: []byte("{ \"body\": \"{ \\\"data\\\": { \\\"result\\\": true }}\"}"),
		},
	}
	client := Client{
		c: &mock,
	}

	vars := map[string]interface{}{
		"var": "value",
	}

	res, err := client.Gql("some_lambda:status/some/path", MOCK_MUTATION, vars)
	if err != nil {
		t.Fatal("Unexpected test Error", err)
	}
	if !mock.hasBeenCalled {
		t.Fatal("Mock Invoke never called")
	}
	if *mock.payload.FunctionName != "some_lambda:status" {
		t.Fatal("Did not use correct function name", mock.payload.FunctionName)
	}

	if !(*res)["result"].(bool) {
		t.Fatal("Did not return data", *res)
	}
	mock.response = &lambda.InvokeOutput{
		Payload: []byte("{ \"body\": \"{\\\"errors\\\": [{ \\\"message\\\": \\\"error message\\\"}] }\" }"),
	}
	res, err = client.Gql("some_lambda:status/some/path", MOCK_MUTATION, vars)
	if res != nil {
		t.Fatal("Unexpected return value", *res)
	}
	if err == nil {
		t.Fatal("Should have returned error value")
	}
	if err.Error() != "error message" {
		t.Fatal("Did not return needed error message", err.Error())
	}
}

func testParseUri(t *testing.T) {
	functionName, path, err := parseUri("some_lambda:status/some/path")
	if err != nil {
		t.Fatal("Unexpected error", err)
	}
	if *functionName != "some_lambda:status" {
		t.Fatal("Did not parse function name right", *functionName)
	}
	if *path != "/some/path" {
		t.Fatal("Did not parse path right", *path)
	}

	functionName, path, err = parseUri("some_lambda:status.invalid_path")

	if err == nil {
		t.Fatal("Expected an error")
	}
}