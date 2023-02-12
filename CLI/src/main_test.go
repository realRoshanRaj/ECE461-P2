package main

import (
	"testing"
	// dep "CLI/node_modules"
	// dep "CLI/dependencies"
	// "context"
	// "math"
	// "strconv"
	// "time"
	"fmt"
	// "log"
	"net/http"
	// "net/http/httputil"
	// "bufio"
	"os"
	"strings"

	// "github.com/joho/godotenv"
	// "github.com/machinebox/graphql"
)

// func TestInit(t *testing.T) {
// 	os.Setenv("GITHUB_TOKEN", "test_token")
// 	defer os.Unsetenv("GITHUB_TOKEN")

// 	godotenv.Load(".env")
// 	token := os.Getenv("GITHUB_TOKEN")

// 	if token != "test_token" { t.Errorf("Expected token to be 'test_token', but got %s", token) }
// 	if repos == nil { t.Errorf("Repos dne")	}
// }

func TestConvertUrl(t *testing.T) {
	// godotenv.Load(".env")
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
	// godotenv.Load(".env")
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
	// godotenv.Load(".env")
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

func TestGraphqlFunc(t *testing.T) {
	// godotenv.Load(".env")
	token = os.Getenv("GITHUB_TOKEN")

	graphql_func("cloudinary", "cloudinary_npm", token)
}

func TestRespDataql1(t *testing.T) {
	data := respDataql1{
		Repository: struct {
			Issues struct { TotalCount int }
			PullRequests struct { TotalCount int }
			Upcase struct { Text string }
			Downcase struct { Text string }
			Capcase struct { Text string }
			Expcase struct { Text string }
			Commits struct { History struct { TotalCount int } }
		}{ Issues: struct { TotalCount int } { TotalCount: 10, },
			PullRequests: struct { TotalCount int }{ TotalCount: 5, },
			Upcase: struct { Text string } { Text: "README CONTENT", },
			Downcase: struct { Text string } { Text: "readme content", },
			Capcase: struct { Text string }{ Text: "This is the content of Readme.md", },
			Expcase: struct { Text string }{ Text: "This is the content of readme.markdown", },
			Commits: struct { History struct { TotalCount int } } {
				History: struct { TotalCount int }{	TotalCount: 20,	}, }, },
	}

	if data.Repository.Issues.TotalCount != 10 {
		t.Errorf("Expected Issues TotalCount to be 10, but got %d", data.Repository.Issues.TotalCount)
	}

	if data.Repository.PullRequests.TotalCount != 5 {
		t.Errorf("Expected pull requests to be 5, got %d", data.Repository.PullRequests.TotalCount)
	}
}

func TestMain(t *testing.T) {
	var totalTests, passedTests int
	tests := []struct {
		name string
		f    func(*testing.T)
	}{
		{"Test Case 1", TestCase1},
		{"Test Case 2", TestCase2},
		{"Test Case 3", TestCase3},
		{"Test Case 4", TestCase4},
		{"Test Case 5", TestCase5},
		{"Test Case 6", TestCase6},
		{"Test Case 7", TestCase7},
		{"Test Case 8", TestCase8},
		{"Test Case 9", TestCase9},
		{"Test Case 10", TestCase10},
		{"Test Case 11", TestCase11},
		{"Test Case 12", TestCase12},
		{"Test Case 13", TestCase13},
		{"Test Case 14", TestCase14},
		{"Test Case 15", TestCase15},
		{"Test Case 16", TestCase16},
		{"Test Case 17", TestCase17},
		{"Test Case 18", TestCase18},
		{"Test Case 19", TestCase19},
		{"Test Case 20", TestCase20},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			totalTests++
			test.f(t)
			passedTests++
		})
	}
	fmt.Printf("\n%d/%d test cases passed. %.2f%%\n", passedTests, totalTests, float64(passedTests)/float64(totalTests)*100)
}

func TestCase1(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"run", "testdata.txt"}

	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	f, err := os.Create("testdata.txt")
	if (err != nil) { t.Errorf("Could not create testdata file") }
	f.WriteString("https://github.com/lodash/lodash")
	defer f.Close()

	main()
	os.Remove("testdata.txt")
}

func TestCase2(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"run", "testdata.txt"}

	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	f, err := os.Create("testdata.txt")
	if (err != nil) { t.Errorf("Could not create testdata file") }
	f.WriteString("https://github.com/lulululbj/wanandroid")
	defer f.Close()

	main()
	os.Remove("testdata.txt")
}

func TestCase3(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"run", "testdata.txt"}

	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	f, err := os.Create("testdata.txt")
	if (err != nil) { t.Errorf("Could not create testdata file") }
	f.WriteString("https://github.com/diyhue/diyHue")
	defer f.Close()

	main()
	os.Remove("testdata.txt")
}

func TestCase4(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"run", "testdata.txt"}

	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	f, err := os.Create("testdata.txt")
	if (err != nil) { t.Errorf("Could not create testdata file") }
	f.WriteString("https://github.com/Jounce/Surge")
	defer f.Close()

	main()
	os.Remove("testdata.txt")
}

func TestCase5(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"run", "testdata.txt"}

	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	f, err := os.Create("testdata.txt")
	if (err != nil) { t.Errorf("Could not create testdata file") }
	f.WriteString("https://github.com/ReactTraining/react-router")
	defer f.Close()

	main()
	os.Remove("testdata.txt")
}

func TestCase6(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"run", "testdata.txt"}

	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	f, err := os.Create("testdata.txt")
	if (err != nil) { t.Errorf("Could not create testdata file") }
	f.WriteString("https://github.com/mabe02/lanterna")
	defer f.Close()

	main()
	os.Remove("testdata.txt")
}

func TestCase7(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"run", "testdata.txt"}

	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	f, err := os.Create("testdata.txt")
	if (err != nil) { t.Errorf("Could not create testdata file") }
	f.WriteString("https://github.com/nidem/kerberoast")
	defer f.Close()

	main()
	os.Remove("testdata.txt")
}

func TestCase8(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"run", "testdata.txt"}

	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	f, err := os.Create("testdata.txt")
	if (err != nil) { t.Errorf("Could not create testdata file") }
	f.WriteString("https://github.com/AssemblyScript/prototype")
	defer f.Close()

	main()
	os.Remove("testdata.txt")
}

func TestCase9(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"run", "testdata.tx"}

	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	f, err := os.Create("testdata.txt")
	if (err != nil) { t.Errorf("Could not create testdata file") }
	f.WriteString("https://github.com/panjf2000/gne")
	defer f.Close()

	main()
	os.Remove("testdata.txt")
}

func TestCase10(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"run", "testdata.txt"}

	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	f, err := os.Create("testdata.txt")
	if (err != nil) { t.Errorf("Could not create testdata file") }
	f.WriteString("https://github.com/xflux-gui/fluxgui")
	defer f.Close()

	main()
	os.Remove("testdata.txt")
}

func TestCase11(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"run", "testdata.txt"}

	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	f, err := os.Create("testdata.txt")
	if (err != nil) { t.Errorf("Could not create testdata file") }
	f.WriteString("https://github.com/Cartucho/mAP")
	defer f.Close()

	main()
	os.Remove("testdata.txt")
}

func TestCase12(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"run", "testdata.txt"}

	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	f, err := os.Create("testdata.txt")
	if (err != nil) { t.Errorf("Could not create testdata file") }
	f.WriteString("https://github.com/dhamaniasad/HeadlessBrowsers")
	defer f.Close()

	main()
	os.Remove("testdata.txt")
}

func TestCase13(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"run", "testdata.txt"}

	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	f, err := os.Create("testdata.txt")
	if (err != nil) { t.Errorf("Could not create testdata file") }
	f.WriteString("https://github.com/madrobby/keymaster")
	defer f.Close()

	main()
	os.Remove("testdata.txt")
}

func TestCase14(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"run", "testdata.txt"}

	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	f, err := os.Create("testdata.txt")
	if (err != nil) { t.Errorf("Could not create testdata file") }
	f.WriteString("https://github.com/rougier/matplotlib-cheatsheet")
	defer f.Close()

	main()
	os.Remove("testdata.txt")
}

func TestCase15(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"run", "testdata.txt"}

	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	f, err := os.Create("testdata.txt")
	if (err != nil) { t.Errorf("Could not create testdata file") }
	f.WriteString("https://github.com/lodash/lodash")
	defer f.Close()

	main()
	os.Remove("testdata.txt")
}

func TestCase16(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"run", "testdata.txt"}

	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	f, err := os.Create("testdata.txt")
	if (err != nil) { t.Errorf("Could not create testdata file") }
	f.WriteString("https://github.com/cruip/tailwind-landing-page-template")
	defer f.Close()

	main()
	os.Remove("testdata.txt")
}

func TestCase17(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"run", "testdata.txt"}

	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	f, err := os.Create("testdata.txt")
	if (err != nil) { t.Errorf("Could not create testdata file") }
	f.WriteString("https://github.com/mike4192/spotMicro")
	defer f.Close()

	main()
	os.Remove("testdata.txt")
}

func TestCase18(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"run", "testdata.txt"}

	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	f, err := os.Create("testdata.txt")
	if (err != nil) { t.Errorf("Could not create testdata file") }
	f.WriteString("https://github.com/unrolled/render")
	defer f.Close()

	main()
	os.Remove("testdata.txt")
}

func TestCase19(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"run", "testdata.txt"}

	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	f, err := os.Create("testdata.txt")
	if (err != nil) { t.Errorf("Could not create testdata file") }
	f.WriteString("https://github.com/matcornic/hugo-theme-learn")
	defer f.Close()

	main()
	os.Remove("testdata.txt")
}

func TestCase20(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"run", "testdata.txt"}

	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	f, err := os.Create("testdata.txt")
	if (err != nil) { t.Errorf("Could not create testdata file") }
	f.WriteString("https://github.com/gomatcha/matcha")
	defer f.Close()

	main()
	os.Remove("testdata.txt")
}

/* COVERED BY GRAPHQL FUNC TEST
func TestStoreLog(t *testing.T) {
	empty := []byte {};
	//data := []byte("Test data")
	header := "Test header"
	filename := "test.log"

	err := storeLog(filename, empty , header, true)
	if err != nil {
		t.Errorf("Error storing log data: %v", err)
	}

	os.Remove(filename)
}
*/
