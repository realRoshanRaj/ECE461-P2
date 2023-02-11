package main

import "testing"

import (
	dep "CLI/dependencies"
	"testing"
	"os"

	"github.com/joho/dotenv"

	// UNCOMMENT AS YOU NEED THEM
	// "context"
	// "math"
	// "strconv"
	// "time"

	// // json "encoding/json"

	// "flag"
	// "fmt"

	// // "io/ioutil"
	// "log"
	// "net/http"
	// "net/http/httputil"

	// "bufio"
	// "os/exec"
	// "strings"

	// "github.com/machinebox/graphql"
)

func TestInit(t *testing.T) {
	os.Setenv("GITHUB_TOKEN", "test_token")
	defer os.Unsetenv("GITHUB_TOKEN")

	godotenv.Load(".env")
	token := os.Getenv("GITHUB_TOKEN")

	if token != "test_token" {
		t.Errorf("Expected token to be 'test_token', but got %s", token)
	}

	if repos == nil {
		t.Errorf("Expected repos to be initialized, but got nil")
	}
}