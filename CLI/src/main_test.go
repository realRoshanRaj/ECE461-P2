package main

import (
	"testing"
	dep "CLI/node_modules"
	// dep "CLI/dependencies"
	// "context"
	// "math"
	// "strconv"
	// "time"
	// "fmt"
	// "log"
	// "net/http"
	// "net/http/httputil"
	// "bufio"
	"os"
	// "os/exec"
	// "strings"

	// These are dependencies must be installed with go get make sure in makefile
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

// DOES NOT RUN, ASK ABOUT PATH TO NODE?
func TestConvertUrl(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"https://www.npmjs.com", "https://github.com"},
		{"https://www.google.com", "https://www.google.com"},
	}

	for _, test := range tests {
		url := test.input
		convertUrl(&url)
		if url != test.expected {
			t.Errorf("Expected %s, but got %s", test.expected, url)
		}
	}
}
