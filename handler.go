package lambdahandler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"

	core "github.com/jacob-ebey/graphql-core"
)

const (
	ContentTypeJSON           = "application/json"
	ContentTypeGraphQL        = "application/graphql"
	ContentTypeFormURLEncoded = "application/x-www-form-urlencoded"
)

// a workaround for getting`variables` as a JSON string
type graphQLRequestCompatibility struct {
	Query         string `json:"query" url:"query" schema:"query"`
	OperationName string `json:"operationName" url:"operationName" schema:"operationName"`
	Variables     string `json:"variables" url:"variables" schema:"variables"`
}

func getFromForm(values map[string]string) *core.GraphQLRequest {
	query, ok := values["query"]
	if ok && query != "" {
		// get variables map
		variables := map[string]interface{}{}
		variablesStr, ok := values["variables"]
		if ok && variablesStr != "" {
			json.Unmarshal([]byte(variablesStr), &variables)
		}

		operationName, _ := values["operationName"]

		return &core.GraphQLRequest{
			Query:         query,
			Variables:     variables,
			OperationName: operationName,
		}
	}

	return nil
}

func newGraphQLRequest(r events.APIGatewayProxyRequest) core.GraphQLRequest {
	if reqOpt := getFromForm(r.QueryStringParameters); reqOpt != nil {
		return *reqOpt
	}

	contentTypeStr, _ := r.Headers["Content-Type"]
	contentTypeTokens := strings.Split(contentTypeStr, ";")
	contentType := contentTypeTokens[0]

	switch contentType {
	case ContentTypeGraphQL:
		return core.GraphQLRequest{
			Query: string(r.Body),
		}
	case ContentTypeFormURLEncoded:
		// TODO: Figure out how to handle Form Encoded
		return core.GraphQLRequest{}
	case ContentTypeJSON:
		fallthrough
	default:
		opts := core.GraphQLRequest{}
		err := json.Unmarshal([]byte(r.Body), &opts)
		if err != nil {
			// Probably `variables` was sent as a string instead of an object.
			// So, we try to be polite and try to parse that as a JSON string
			var optsCompatible graphQLRequestCompatibility
			json.Unmarshal([]byte(r.Body), &optsCompatible)
			json.Unmarshal([]byte(optsCompatible.Variables), &opts.Variables)
		}
		return opts
	}
}

type GraphQLHttpLambdaHandler struct {
	Executor core.GraphQLExecutor
}

func (handler *GraphQLHttpLambdaHandler) LambdaHandler(r events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	request := newGraphQLRequest(r)
	result := handler.Executor.Execute(context.TODO(), request)

	json, err := json.Marshal(result)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
		}, nil
	}

	return events.APIGatewayProxyResponse{
		Body:       string(json),
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json; charset=utf-8",
		},
	}, nil
}
