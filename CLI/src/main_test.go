package main

import (
	"testing"
	// dep "CLI/node_modules"
	// dep "CLI/dependencies"
	// "context"
	// "math"
	// "strconv"
	// "time"
	// "fmt"
	// "log"
	"net/http"
	// "net/http/httputil"
	// "bufio"
	"os"
	"strings"

	"github.com/joho/godotenv"
	// "github.com/machinebox/graphql"
)

func TestInit(t *testing.T) {
	os.Setenv("GITHUB_TOKEN", "test_token")
	defer os.Unsetenv("GITHUB_TOKEN")

	godotenv.Load(".env")
	token := os.Getenv("GITHUB_TOKEN")

	if token != "test_token" { t.Errorf("Expected token to be 'test_token', but got %s", token) }
	if repos == nil { t.Errorf("Repos dne")	}
}

func TestConvertUrl(t *testing.T) {
	godotenv.Load(".env")
	token = os.Getenv("GITHUB_TOKEN")

	tests := []struct {
		input    string
		expected string
	}{
		// exe.Command does not work when called from here
		//{"https://www.npmjs.com", "git://github.com"},
		{"https://www.google.com", "https://www.google.com"},
	}

	for _, test := range tests {
		input := test.input
		expected := test.expected
		convertUrl(&input)
		if input != expected {
			t.Errorf("convert(%q); Expected %s, but got %s", test.input, expected, input)
		}
	}
}

func TestGetRepoResponse(t *testing.T) {
	godotenv.Load(".env")
	token = os.Getenv("GITHUB_TOKEN")

	resp := getRepoResponse("https://github.com/nullivex/nodist")

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Unexpected response status: %d", resp.StatusCode)
	}

	if h := resp.Request.Header.Get("Authorization"); h != "Bearer " + token {
		t.Fatalf("Unexpected Authorization header value: %q", h)
	}

	if !strings.HasPrefix(resp.Request.URL.String(), "https://api.github.com/repos/") {
		t.Fatalf("Unexpected URL format: %q", resp.Request.URL.String())
	}
}


func TestGetContributorResponse(t *testing.T) {
	godotenv.Load(".env")
	token = os.Getenv("GITHUB_TOKEN")

	testCases := []struct {
		name         string
		httpUrl      string
		expectedCode int
	}{
		{
			name:         "Successful response",
			httpUrl:      "https://github.com/nullivex/nodist",
			expectedCode: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res := getContributorResponse(tc.httpUrl)

			if res.StatusCode != tc.expectedCode {
				t.Fatalf("Unexpected response status: %d", res.StatusCode)
			}
		})
	}
}
