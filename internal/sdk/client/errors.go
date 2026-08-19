package client

import (
	"encoding/json"
	"fmt"
	"strings"
)

type GraphQLError struct {
	Message    string                 `json:"message"`
	Path       []interface{}          `json:"path"` // GraphQL path elements are strings or list indices (ints)
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

type errorMapping struct {
	Message         string // exact match on GraphQLError.Message
	FriendlyMessage string // user-friendly message with %s for resource name
}

var knownErrors = []errorMapping{
	{Message: "conflict", FriendlyMessage: "environment '%s' already exists"},
	{Message: "Invalid API token", FriendlyMessage: "invalid API token, please verify your credentials. You can get an Anywhere API token at https://acm.altinity.cloud/account"},
	{Message: "hosted env management not enabled", FriendlyMessage: "Altinity-hosted environments are not enabled for your account, ask Altinity support to enable them"},
}

// FormatError falls back to the parsed GraphQL messages rather than the raw JSON string.
func FormatError(err error, resourceName string) string {
	// ParseError reports (nil, nil) for a nil error, which every path below dereferences.
	if err == nil {
		return ""
	}

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

// errorPrefix infers "Validation Error" from a mutation path when there is no extensions code.
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

	for _, p := range gqlError.Path {
		s, ok := p.(string)
		if !ok {
			continue
		}
		lp := strings.ToLower(s)
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
