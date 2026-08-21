package client

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Yamashou/gqlgenc/clientv2"
	"github.com/vektah/gqlparser/v2/ast"
)

type GraphQLError struct {
	Message    string
	Path       []interface{} // GraphQL path elements are strings or list indices (ints)
	Extensions map[string]interface{}
}

func (e GraphQLError) Error() string {
	return fmt.Sprintf("GraphQL Error: %s", e.Message)
}

// ClientError is the projection of clientv2.ErrorResponse the helpers below work
// on, so nothing outside this file depends on the SDK's error shape.
type ClientError struct {
	NetworkError  *clientv2.HTTPError
	GraphqlErrors []GraphQLError
}

// ParseError resolves the typed error clientv2 returns. Transport failures never
// reach a response, so clientv2 wraps them plainly and there is nothing to read;
// those are reported as an error rather than an empty ClientError.
func ParseError(err error) (*ClientError, error) {
	if err == nil {
		return nil, nil
	}

	var errResp *clientv2.ErrorResponse
	if !errors.As(err, &errResp) {
		return nil, fmt.Errorf("not a GraphQL error response: %w", err)
	}

	parsed := &ClientError{NetworkError: errResp.NetworkError}
	if errResp.GqlErrors != nil {
		for _, gqlErr := range *errResp.GqlErrors {
			if gqlErr == nil {
				continue
			}
			parsed.GraphqlErrors = append(parsed.GraphqlErrors, GraphQLError{
				Message:    gqlErr.Message,
				Path:       pathElements(gqlErr.Path),
				Extensions: gqlErr.Extensions,
			})
		}
	}

	return parsed, nil
}

// pathElements flattens ast.PathName and ast.PathIndex to the plain string and int
// that errorPrefix type-asserts on.
func pathElements(path ast.Path) []interface{} {
	var elements []interface{}
	for _, element := range path {
		switch v := element.(type) {
		case ast.PathName:
			elements = append(elements, string(v))
		case ast.PathIndex:
			elements = append(elements, int(v))
		default:
			elements = append(elements, element)
		}
	}

	return elements
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

	if parsedError.NetworkError != nil {
		return formatNetworkError(parsedError.NetworkError)
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

func formatNetworkError(networkError *clientv2.HTTPError) string {
	if networkError.Message == "" {
		return fmt.Sprintf("Network error: HTTP %d", networkError.Code)
	}

	return fmt.Sprintf("Network error: %s", networkError.Message)
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
