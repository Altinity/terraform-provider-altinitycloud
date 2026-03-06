package client

import (
	"errors"
	"testing"
)

func TestFormatError_ValidationError(t *testing.T) {
	rawErr := errors.New(`{"networkErrors":null,"graphqlErrors":[{"message":"iceberg: iceberg: catalog \"ianaya89-tf-test\": table \"hola.ok\": invalid path: \"hola/\"","path":["updateAWSEnv"]}]}`)
	got := FormatError(rawErr, "ianaya89-tf-test")
	want := `Validation Error: iceberg: iceberg: catalog "ianaya89-tf-test": table "hola.ok": invalid path: "hola/"`
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestFormatError_NotFoundWithExtension(t *testing.T) {
	rawErr := errors.New(`{"networkErrors":null,"graphqlErrors":[{"message":"env not found","path":["getAWSEnv"],"extensions":{"code":"NOT_FOUND"}}]}`)
	got := FormatError(rawErr, "test")
	want := "Not Found: env not found"
	if got != want {
		t.Errorf("got: %s, want: %s", got, want)
	}
}

func TestFormatError_MultipleErrors(t *testing.T) {
	rawErr := errors.New(`{"networkErrors":null,"graphqlErrors":[{"message":"field X is required","path":["createAWSEnv"]},{"message":"field Y is invalid","path":["createAWSEnv"]}]}`)
	got := FormatError(rawErr, "test")
	want := "Validation Error: field X is required\nValidation Error: field Y is invalid"
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestFormatError_KnownError(t *testing.T) {
	rawErr := errors.New(`{"networkErrors":null,"graphqlErrors":[{"message":"conflict","path":["createAWSEnv"]}]}`)
	got := FormatError(rawErr, "my-env")
	want := "environment 'my-env' already exists"
	if got != want {
		t.Errorf("got: %s, want: %s", got, want)
	}
}

func TestFormatError_NonGraphQLError(t *testing.T) {
	rawErr := errors.New("connection refused")
	got := FormatError(rawErr, "test")
	want := "connection refused"
	if got != want {
		t.Errorf("got: %s, want: %s", got, want)
	}
}

func TestFormatError_QueryPathFallback(t *testing.T) {
	rawErr := errors.New(`{"networkErrors":null,"graphqlErrors":[{"message":"something went wrong","path":["getAWSEnv"]}]}`)
	got := FormatError(rawErr, "test")
	want := "Error: something went wrong"
	if got != want {
		t.Errorf("got: %s, want: %s", got, want)
	}
}
