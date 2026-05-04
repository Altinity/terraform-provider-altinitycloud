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
	if parsedError == nil {
		return false, nil
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
// If the error is not recognized, it falls back to a clean representation
// of the GraphQL error messages instead of the raw JSON string.
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

	// Fallback: extract clean error messages from GraphQL errors
	// and classify them based on the extension code or path context.
	var messages []string
	for _, gqlError := range parsedError.GraphqlErrors {
		prefix := errorPrefix(gqlError)
		messages = append(messages, fmt.Sprintf("%s: %s", prefix, gqlError.Message))
	}
	if len(messages) > 0 {
		return strings.Join(messages, "\n")
	}

	if parsedError.NetworkErrors != nil {
		return formatNetworkErrors(parsedError.NetworkErrors)
	}

	return err.Error()
}

// errorPrefix returns a human-readable error category based on the GraphQL
// error extensions code. When no code is present, it infers "Validation Error"
// for mutation paths and defaults to "Error" otherwise.
func errorPrefix(gqlError GraphQLError) string {
	if code, ok := gqlError.Extensions["code"]; ok {
		switch code {
		case "NOT_FOUND":
			return "Not Found"
		case "CONFLICT":
			return "Conflict"
		case "FORBIDDEN":
			return "Forbidden"
		case "UNAUTHORIZED":
			return "Unauthorized"
		default:
			return fmt.Sprintf("%v", code)
		}
	}

	// No extension code: infer from mutation path (create/update/delete).
	for _, p := range gqlError.Path {
		lp := strings.ToLower(p)
		if strings.HasPrefix(lp, "create") || strings.HasPrefix(lp, "update") || strings.HasPrefix(lp, "delete") {
			return "Validation Error"
		}
	}

	return "Error"
}

func formatNetworkErrors(networkErrors interface{}) string {
	switch v := networkErrors.(type) {
	case string:
		return fmt.Sprintf("Network error: %s", v)
	case map[string]interface{}:
		if msg, ok := v["message"]; ok {
			return fmt.Sprintf("Network error: %v", msg)
		}
		raw, _ := json.Marshal(v)
		return fmt.Sprintf("Network error: %s", raw)
	default:
		raw, _ := json.Marshal(v)
		return fmt.Sprintf("Network error: %s", raw)
	}
}

func IsActiveClustersError(err error) (bool, error) {
	parsedError, parseErr := ParseError(err)
	if parseErr != nil {
		return false, parseErr
	}
	if parsedError == nil {
		return false, nil
	}

	for _, gqlError := range parsedError.GraphqlErrors {
		if code, ok := gqlError.Extensions["code"]; ok && code == "CONFLICT" {
			return strings.Contains(gqlError.Message, "forceDestroyClusters=true"), nil
		}
	}

	return false, nil
}
