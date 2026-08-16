package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/benelog/uploader/internal/auth"
	"github.com/benelog/uploader/internal/uploader"
)

const (
	testUser     = "tester"
	testPassword = "s3cr3t"
)

func TestBuiltHandlerRequiresCredentialsByDefault(t *testing.T) {
	rec := request(t, testCreds(), "/form.html", "", "")

	// The details of the challenge (header, realm) are covered by the auth
	// package's own tests; here only the wiring matters.
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestBuiltHandlerServesTheFormWithCredentials(t *testing.T) {
	rec := request(t, testCreds(), "/form.html", testUser, testPassword)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "ready!") {
		t.Error("the upload form should be served once authenticated")
	}
}

func TestNoAuthServesWithoutCredentials(t *testing.T) {
	rec := request(t, nil, "/form.html", "", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestAuthPolicyPrintsAGeneratedPassword(t *testing.T) {
	creds, err := auth.Resolve("", "")
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	var out bytes.Buffer
	printAuthPolicy(&out, &creds, 8080)

	banner := out.String()
	for _, want := range []string{creds.Username, creds.Password, "randomly generated", "--no-auth"} {
		if !strings.Contains(banner, want) {
			t.Errorf("startup banner does not mention %q:\n%s", want, banner)
		}
	}
}

func TestAuthPolicyKeepsAGivenPasswordOffTheConsole(t *testing.T) {
	creds, err := auth.Resolve("alice", "hunter2")
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	var out bytes.Buffer
	printAuthPolicy(&out, &creds, 8080)

	if strings.Contains(out.String(), "hunter2") {
		t.Errorf("a password given with --password was echoed:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "alice") {
		t.Errorf("the username should still be shown:\n%s", out.String())
	}
}

func TestAuthPolicyWarnsWhenAuthIsDisabled(t *testing.T) {
	var out bytes.Buffer
	printAuthPolicy(&out, nil, 8080)

	banner := out.String()
	if !strings.Contains(banner, "DISABLED") || !strings.Contains(banner, "--no-auth") {
		t.Errorf("the banner should warn that authentication is off:\n%s", banner)
	}
}

func testCreds() *auth.Credentials {
	return &auth.Credentials{Username: testUser, Password: testPassword}
}

// request drives a request through exactly the handler chain main builds, with
// nil creds standing for --no-auth. An empty username means "send no
// credentials at all".
func request(t *testing.T, creds *auth.Credentials, path, username, password string) *httptest.ResponseRecorder {
	t.Helper()
	handler, err := uploader.NewHandler()
	if err != nil {
		t.Fatalf("NewHandler failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, path, nil)
	if username != "" {
		req.SetBasicAuth(username, password)
	}
	rec := httptest.NewRecorder()
	buildHandler(handler.Routes(), creds).ServeHTTP(rec, req)
	return rec
}
