package lambdahandler_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/graphql-go/graphql"

	core "github.com/jacob-ebey/graphql-core"

	lambdahandler "github.com/jacob-ebey/graphql-lambdahandler"
	"github.com/jacob-ebey/graphql-lambdahandler/schemas"
)

func TestWrappedErrorDefaultMessage(t *testing.T) {
	handler := lambdahandler.GraphQLHttpLambdaHandler{
		Executor: core.GraphQLExecutor{
			Schema: schemas.PingPongSchema,
		},
	}

	query := "query Test { ping }"

	jsonRequest, err := json.Marshal(core.GraphQLRequest{
		Query:         query,
		OperationName: "Test",
		Variables: map[string]interface{}{
			"echo": "test",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	jsonStringVariablesRequest, err := json.Marshal(map[string]interface{}{
		"query":         query,
		"operationName": "Test",
		"variables":     `{ "echo": "test" }`,
	})
	if err != nil {
		t.Fatal(err)
	}

	cases := []events.APIGatewayProxyRequest{
		events.APIGatewayProxyRequest{
			Headers: map[string]string{
				"Content-Type": lambdahandler.ContentTypeGraphQL,
			},
			Body: "query { ping }",
		},
		events.APIGatewayProxyRequest{
			QueryStringParameters: map[string]string{
				"query":         query,
				"operationName": "Test",
				"variables":     `{ "echo": "test" }`,
			},
		},
		events.APIGatewayProxyRequest{
			Headers: map[string]string{
				"Content-Type": lambdahandler.ContentTypeJSON,
			},
			Body: string(jsonRequest),
		},
		events.APIGatewayProxyRequest{
			Body: string(jsonRequest),
		},
		events.APIGatewayProxyRequest{
			Body: string(jsonStringVariablesRequest),
		},
	}

	for _, request := range cases {
		response, err := handler.LambdaHandler(request)

		if err != nil {
			t.Fatal(err)
		}

		if response.StatusCode != http.StatusOK {
			t.Fatal("Status code was not 200.")
		}

		result := graphql.Result{}
		if err := json.Unmarshal([]byte(response.Body), &result); err != nil {
			t.Fatal(err)
		}

		if result.HasErrors() {
			t.Fatal(result.Errors)
		}
	}
}
