package client

import (
	"encoding/json"
	"fmt"
	"strings"
)

type GraphQLError struct {
	Message    string                 `json:"message"`
	Path       []string               `json:"path"`
	Extensions map[string]interface{} `json:"extensions"`
}

func (e GraphQLError) Error() string {
	return fmt.Sprintf("GraphQL Error: %s", e.Message)
}

type ClientError struct {
	NetworkErrors []interface{}  `json:"networkErrors"`
	GraphqlErrors []GraphQLError `json:"graphqlErrors"`
}

func ParseError(err error) (*ClientError, error) {
	if err == nil {
		return nil, nil
	}

	var errResp ClientError
	if jsonErr := json.Unmarshal([]byte(err.Error()), &errResp); jsonErr != nil {
		return nil, fmt.Errorf("error parsing: %v", jsonErr)
	}

	return &errResp, nil
}

func IsNotFoundError(err error) (bool, error) {
	parsedError, parseErr := ParseError(err)
	if parseErr != nil {
		return false, parseErr
	}

	for _, gqlError := range parsedError.GraphqlErrors {
		if code, ok := gqlError.Extensions["code"]; ok && code == "NOT_FOUND" {
			return true, nil
		}
	}

	return false, nil
}

func IsActiceClustersError(err error) (bool, error) {
	parsedError, parseErr := ParseError(err)
	if parseErr != nil {
		return false, parseErr
	}

	for _, gqlError := range parsedError.GraphqlErrors {
		if code, ok := gqlError.Extensions["code"]; ok && code == "CONFLICT" {
			return strings.Contains(gqlError.Message, "forceDestroyClusters=true"), nil
		}
	}

	return false, nil
}
