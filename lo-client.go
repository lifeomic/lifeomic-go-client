package lo_client

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
)

type payload struct {
	Headers               map[string]string `json:"headers"`
	Path                  string            `json:"path"`
	HttpMethod            string            `json:"httpMethod"`
	QueryStringParameters map[string]string `json:"queryStringParameters"`
	Body                  string            `json:"body"`
}

type policy struct {
	Rules map[string]bool `json:"rules"`
}

type responsePayload struct {
	Body string `json:"body"`
}

type responseBody struct {
	Data   map[string]interface{} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

type Invoker interface {
	Invoke(context.Context, *lambda.InvokeInput, ...func(*lambda.Options)) (*lambda.InvokeOutput, error)
}

type Client struct {
	c       Invoker
	account string
	user    string
	rules   map[string]bool
}

func (client *Client) buildGqlQuery(path string, query string, variables map[string]interface{}) []byte {
	type Body struct {
		Query     string                 `json:"query"`
		Variables map[string]interface{} `json:"variables"`
	}
	policy, _ := json.Marshal(&policy{
		Rules: client.rules,
	})
	body, _ := json.Marshal(&Body{Query: query, Variables: variables})
	payload := &payload{
		Headers:               map[string]string{"LifeOmic-Account": client.account, "LifeOmic-User": client.user, "content-type": "application/json", "LifeOmic-Policy": string(policy)},
		HttpMethod:            "POST",
		QueryStringParameters: map[string]string{},
		Path:                  path,
		Body:                  string(body),
	}
	bytes, err := json.Marshal(payload)
	if err != nil {
		log.Fatalf("Failed to marshall payload %v", err)
	}
	return bytes
}

func parseUri(uri string) (*string, *string, error) {
	index := strings.IndexAny(uri, "/")
	if index == -1 {
		return nil, nil, errors.New("Invalid URL provided")
	}
	functionName := uri[0:index]
	path := uri[index:]
	return &functionName, &path, nil
}

func (client *Client) Gql(uri string, query string, variables map[string]interface{}) (*map[string]interface{}, error) {
	functionName, path, err := parseUri(uri)
	if err != nil {
		return nil, err
	}
	// MP_ARN := "marketplace-service:deployed"
	resp, err := client.c.Invoke(context.Background(), &lambda.InvokeInput{
		FunctionName: functionName,
		Payload:      client.buildGqlQuery(*path, query, variables),
	})

	if err != nil {
		return nil, err
	}
	var payload responsePayload
	err = json.Unmarshal(resp.Payload, &payload)
	if err != nil {
		return nil, err
	}

	var body responseBody
	err = json.Unmarshal([]byte(payload.Body), &body)
	if err != nil {
		return nil, err
	}
	if len(body.Errors) > 0 {
		return nil, errors.New(body.Errors[0].Message)
	}
	return &body.Data, nil
}

func BuildClient(account string, user string, rules map[string]bool) (*Client, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, err
	}
	client := Client{c: lambda.NewFromConfig(cfg), user: user, rules: rules, account: account}
	return &client, nil
}
