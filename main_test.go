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
	rec := request(t, false, "/form.html", "", "")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if challenge := rec.Header().Get("WWW-Authenticate"); !strings.HasPrefix(challenge, "Basic ") {
		t.Errorf("WWW-Authenticate = %q, want a Basic challenge", challenge)
	}
}

func TestBuiltHandlerServesTheFormWithCredentials(t *testing.T) {
	rec := request(t, false, "/form.html", testUser, testPassword)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "ready!") {
		t.Error("the upload form should be served once authenticated")
	}
}

func TestNoAuthServesWithoutCredentials(t *testing.T) {
	rec := request(t, true, "/form.html", "", "")

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
	printAuthPolicy(&out, creds, false, 8080)

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
	printAuthPolicy(&out, creds, false, 8080)

	if strings.Contains(out.String(), "hunter2") {
		t.Errorf("a password given with --password was echoed:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "alice") {
		t.Errorf("the username should still be shown:\n%s", out.String())
	}
}

func TestAuthPolicyWarnsWhenAuthIsDisabled(t *testing.T) {
	var out bytes.Buffer
	printAuthPolicy(&out, auth.Credentials{}, true, 8080)

	banner := out.String()
	if !strings.Contains(banner, "DISABLED") || !strings.Contains(banner, "--no-auth") {
		t.Errorf("the banner should warn that authentication is off:\n%s", banner)
	}
}

// request drives a request through exactly the handler chain main builds. An
// empty username means "send no credentials at all".
func request(t *testing.T, authDisabled bool, path, username, password string) *httptest.ResponseRecorder {
	t.Helper()
	handler, err := uploader.NewHandler()
	if err != nil {
		t.Fatalf("NewHandler failed: %v", err)
	}
	creds := auth.Credentials{Username: testUser, Password: testPassword}

	req := httptest.NewRequest(http.MethodGet, path, nil)
	if username != "" {
		req.SetBasicAuth(username, password)
	}
	rec := httptest.NewRecorder()
	buildHandler(handler.Routes(), creds, authDisabled).ServeHTTP(rec, req)
	return rec
}
