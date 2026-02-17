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
	NetworkErrors interface{}    `json:"networkErrors"`
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

// errorMapping defines a known error pattern and its user-friendly message template.
type errorMapping struct {
	Message         string // exact match on GraphQLError.Message
	FriendlyMessage string // user-friendly message with %s for resource name
}

var knownErrors = []errorMapping{
	{Message: "conflict", FriendlyMessage: "environment '%s' already exists"},
	{Message: "Invalid API token", FriendlyMessage: "invalid API token, please verify your credentials. You can get an Anywhere API token at https://acm.altinity.cloud/account"},
}

// FormatError translates known GraphQL errors into user-friendly messages.
// If the error is not recognized, it falls back to the raw error string.
func FormatError(err error, resourceName string) string {
	parsedError, parseErr := ParseError(err)
	if parseErr != nil {
		return err.Error()
	}

	for _, gqlError := range parsedError.GraphqlErrors {
		for _, mapping := range knownErrors {
			if gqlError.Message == mapping.Message {
				if strings.Contains(mapping.FriendlyMessage, "%s") {
					return fmt.Sprintf(mapping.FriendlyMessage, resourceName)
				}
				return mapping.FriendlyMessage
			}
		}
	}

	return err.Error()
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
