package main

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type Request struct {
	Age    int    `json:"age"`
	Answer string `json:"answer"`
	Result int    `json:"result"`
}

type DBObject struct {
	Timestamp int64  `json:"timestamp"`
	Age       int    `json:"age"`
	Answer    string `json:"answer"`
	Result    int    `json:"result"`
}

var svc *dynamodb.DynamoDB

var CORS = map[string]string{
	"Access-Control-Allow-Headers": "Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token",
	"Access-Control-Allow-Methods": "OPTIONS,POST",
	"Access-Control-Allow-Origin":  "*",
}

func init() {
	sess := session.Must(session.NewSession())
	svc = dynamodb.New(sess)
}

func buildBadRequestError(err string) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		StatusCode: 400,
		Body:       "BadRequest: " + err,
		Headers:    CORS,
	}
}

func buildInternalError(err string) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		StatusCode: 500,
		Body:       "InternalError: " + err,
		Headers:    CORS,
	}
}

func buildOK() events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       "Record Posted",
		Headers:    CORS,
	}
}

func PushDB(item Request) error {
	now := time.Now()
	dbobj := DBObject{
		Timestamp: now.UnixNano(),
		Age:       item.Age,
		Answer:    item.Answer,
		Result:    item.Result,
	}
	av, err := dynamodbattribute.MarshalMap(dbobj)
	if err != nil {
		return err
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String("heartjump-db"),
	}

	_, err = svc.PutItem(input)
	return err
}

func checkItem(item Request) bool {
	if item.Age < 1 || item.Age > 5 {
		return false
	}

	if item.Result < 0 || item.Result > 4 {
		return false
	}

	if len(item.Answer) != 7 {
		return false
	}

	if _, err := strconv.Atoi(item.Answer); err != nil {
		return false
	}
	return true
}

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	item := Request{}
	err := json.Unmarshal([]byte(request.Body), &item)

	if err != nil {
		return buildBadRequestError("Bad JSON body: " + err.Error()), nil
	}

	if !checkItem(item) {
		return buildBadRequestError("Bad Item"), nil
	}

	err = PushDB(item)

	if err != nil {
		return buildInternalError(err.Error()), nil
	}

	return buildOK(), nil
}

func main() {
	lambda.Start(HandleRequest)
}
