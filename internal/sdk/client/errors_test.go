package client

import (
	"errors"
	"fmt"
	"testing"

	"github.com/Yamashou/gqlgenc/clientv2"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// graphQLResponse builds the error clientv2 returns for a response carrying
// GraphQL errors. client_test.go has narrower gqlErr/netErr helpers aimed at the
// retry classification.
func graphQLResponse(errs ...*gqlerror.Error) error {
	list := gqlerror.List(errs)
	return &clientv2.ErrorResponse{GqlErrors: &list}
}

// networkFailure builds the error clientv2 returns when the HTTP status is not OK.
func networkFailure(code int, message string) error {
	return &clientv2.ErrorResponse{NetworkError: &clientv2.HTTPError{Code: code, Message: message}}
}

func mutationErr(message, mutation string) *gqlerror.Error {
	return &gqlerror.Error{Message: message, Path: ast.Path{ast.PathName(mutation)}}
}

func codedErr(message, code string) *gqlerror.Error {
	return &gqlerror.Error{
		Message:    message,
		Extensions: map[string]interface{}{"code": code},
	}
}

func TestFormatError_ValidationError(t *testing.T) {
	err := graphQLResponse(mutationErr(`iceberg: catalog "ianaya89-tf-test": invalid path: "hola/"`, "updateAWSEnv"))

	got := FormatError(err, "ianaya89-tf-test")
	want := `Validation Error: iceberg: catalog "ianaya89-tf-test": invalid path: "hola/"`
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestFormatError_NotFoundWithExtension(t *testing.T) {
	got := FormatError(graphQLResponse(codedErr("env not found", "NOT_FOUND")), "test")
	want := "Not Found: env not found"
	if got != want {
		t.Errorf("got: %s, want: %s", got, want)
	}
}

func TestFormatError_MultipleErrors(t *testing.T) {
	err := graphQLResponse(
		mutationErr("field X is required", "createAWSEnv"),
		mutationErr("field Y is invalid", "createAWSEnv"),
	)

	got := FormatError(err, "test")
	want := "Validation Error: field X is required\nValidation Error: field Y is invalid"
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestFormatError_KnownError(t *testing.T) {
	got := FormatError(graphQLResponse(mutationErr("conflict", "createAWSEnv")), "my-env")
	want := "environment 'my-env' already exists"
	if got != want {
		t.Errorf("got: %s, want: %s", got, want)
	}
}

// A transport failure never reaches a response, so clientv2 wraps it plainly and
// there is nothing to classify.
func TestFormatError_NonGraphQLError(t *testing.T) {
	got := FormatError(errors.New("connection refused"), "test")
	want := "connection refused"
	if got != want {
		t.Errorf("got: %s, want: %s", got, want)
	}
}

// The SDK wraps its error response as it travels up, so the classification has to
// survive wrapping rather than only matching the outermost error.
func TestFormatError_WrappedErrorResponse(t *testing.T) {
	err := fmt.Errorf("executing query: %w", graphQLResponse(codedErr("env not found", "NOT_FOUND")))

	got := FormatError(err, "test")
	want := "Not Found: env not found"
	if got != want {
		t.Errorf("got: %s, want: %s", got, want)
	}
}

func TestFormatError_QueryPathFallback(t *testing.T) {
	err := graphQLResponse(&gqlerror.Error{
		Message: "something went wrong",
		Path:    ast.Path{ast.PathName("getAWSEnv")},
	})

	got := FormatError(err, "test")
	want := "Error: something went wrong"
	if got != want {
		t.Errorf("got: %s, want: %s", got, want)
	}
}

// List indices sit alongside names in a GraphQL path, so the mutation lookup has
// to skip them rather than assume every element is a name.
func TestFormatError_IndexedPath(t *testing.T) {
	err := graphQLResponse(&gqlerror.Error{
		Message: "invalid node type",
		Path:    ast.Path{ast.PathName("createAWSEnv"), ast.PathName("nodeGroups"), ast.PathIndex(2)},
	})

	got := FormatError(err, "test")
	want := "Validation Error: invalid node type"
	if got != want {
		t.Errorf("got: %s, want: %s", got, want)
	}
}

func TestFormatError_NetworkError(t *testing.T) {
	got := FormatError(networkFailure(502, "bad gateway"), "test")
	want := "Network error: bad gateway"
	if got != want {
		t.Errorf("got: %s, want: %s", got, want)
	}
}

func TestFormatError_NetworkErrorWithoutMessage(t *testing.T) {
	got := FormatError(networkFailure(503, ""), "test")
	want := "Network error: HTTP 503"
	if got != want {
		t.Errorf("got: %s, want: %s", got, want)
	}
}

func TestIsNotFoundError_True(t *testing.T) {
	got, err := IsNotFoundError(graphQLResponse(codedErr("env not found", "NOT_FOUND")))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got {
		t.Error("expected true, got false")
	}
}

func TestIsNotFoundError_False(t *testing.T) {
	got, err := IsNotFoundError(graphQLResponse(codedErr("conflict", "CONFLICT")))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got {
		t.Error("expected false, got true")
	}
}

func TestIsNotFoundError_NonGraphQLError(t *testing.T) {
	_, err := IsNotFoundError(errors.New("connection refused"))
	if err == nil {
		t.Error("expected an error for a non-GraphQL failure, got nil")
	}
}

func TestIsNotFoundError_Nil(t *testing.T) {
	got, err := IsNotFoundError(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got {
		t.Error("expected false for nil error")
	}
}

func TestIsActiveClustersError_True(t *testing.T) {
	err := graphQLResponse(codedErr("env has active clusters, use forceDestroyClusters=true", "CONFLICT"))

	got, parseErr := IsActiveClustersError(err)
	if parseErr != nil {
		t.Fatalf("unexpected error: %v", parseErr)
	}
	if !got {
		t.Error("expected true, got false")
	}
}

func TestIsActiveClustersError_ConflictWithoutClusters(t *testing.T) {
	got, err := IsActiveClustersError(graphQLResponse(codedErr("conflict", "CONFLICT")))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got {
		t.Error("expected false, got true")
	}
}

func TestIsActiveClustersError_NotConflict(t *testing.T) {
	got, err := IsActiveClustersError(graphQLResponse(codedErr("env not found", "NOT_FOUND")))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got {
		t.Error("expected false, got true")
	}
}

func TestIsActiveClustersError_NonGraphQLError(t *testing.T) {
	_, err := IsActiveClustersError(errors.New("connection refused"))
	if err == nil {
		t.Error("expected an error for a non-GraphQL failure, got nil")
	}
}

func TestIsActiveClustersError_Nil(t *testing.T) {
	got, err := IsActiveClustersError(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got {
		t.Error("expected false for nil error")
	}
}

func TestParseErrorSkipsNilGraphQLError(t *testing.T) {
	list := gqlerror.List{nil, codedErr("env not found", "NOT_FOUND")}

	parsed, err := ParseError(&clientv2.ErrorResponse{GqlErrors: &list})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(parsed.GraphqlErrors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(parsed.GraphqlErrors))
	}
	if parsed.GraphqlErrors[0].Message != "env not found" {
		t.Errorf("unexpected message: %s", parsed.GraphqlErrors[0].Message)
	}
}
