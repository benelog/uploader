package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	testUser     = "tester"
	testPassword = "s3cr3t"
)

func TestResolveUsesGivenCredentials(t *testing.T) {
	c, err := Resolve("alice", "hunter2")
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	if c.Username != "alice" || c.Password != "hunter2" {
		t.Errorf("got %s/%s, want alice/hunter2", c.Username, c.Password)
	}
	if c.DefaultedUsername || c.GeneratedPassword {
		t.Error("credentials marked as defaulted, want both explicit")
	}
}

func TestResolveDefaultsToOSUserAndRandomPassword(t *testing.T) {
	c, err := Resolve("", "")
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	if c.Username == "" {
		t.Error("username is empty, want the OS login account")
	}
	if !c.DefaultedUsername || !c.GeneratedPassword {
		t.Error("defaults were not flagged as such")
	}
	if len(c.Password) < 16 {
		t.Errorf("generated password %q is shorter than expected", c.Password)
	}
}

func TestResolveGeneratesADifferentPasswordEachTime(t *testing.T) {
	first, err := Resolve("", "")
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	second, err := Resolve("", "")
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	if first.Password == second.Password {
		t.Error("two runs produced the same password")
	}
}

func TestOSUsernameStripsWindowsDomain(t *testing.T) {
	// The stripping logic is inlined in osUsername; check it does not leak a
	// domain prefix on any platform.
	if name := osUsername(); strings.ContainsAny(name, `\/`) {
		t.Errorf("osUsername() = %q, want no domain separator", name)
	}
}

func TestRequestWithoutCredentialsIsUnauthorized(t *testing.T) {
	rec := gated(t, "/form.html", "", "")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	// Without this header the browser never shows a login prompt.
	challenge := rec.Header().Get("WWW-Authenticate")
	if !strings.HasPrefix(challenge, "Basic ") || !strings.Contains(challenge, `realm="uploader"`) {
		t.Errorf("WWW-Authenticate = %q, want a Basic challenge with the uploader realm", challenge)
	}
}

func TestRequestWithWrongPasswordIsUnauthorized(t *testing.T) {
	rec := gated(t, "/form.html", testUser, "wrong")

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestRequestWithWrongUsernameIsUnauthorized(t *testing.T) {
	rec := gated(t, "/form.html", "someone-else", testPassword)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestRequestWithCorrectCredentialsReachesTheHandler(t *testing.T) {
	rec := gated(t, "/form.html", testUser, testPassword)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Body.String() != "reached" {
		t.Errorf("body = %q, want the wrapped handler's answer", rec.Body.String())
	}
}

// The gate sits above the mux, so an unknown path is rejected before the
// routing decides it does not exist: no route can be reached anonymously.
func TestEveryPathIsGated(t *testing.T) {
	for _, path := range []string{"/", "/form.html", "/upload.html", "/nowhere"} {
		if rec := gated(t, path, "", ""); rec.Code != http.StatusUnauthorized {
			t.Errorf("GET %s without credentials = %d, want %d", path, rec.Code, http.StatusUnauthorized)
		}
	}
}

func TestMatchesRejectsEmptyCredentials(t *testing.T) {
	creds := Credentials{Username: testUser, Password: testPassword}
	if creds.Matches("", "") {
		t.Error("empty credentials were accepted")
	}
}

// gated sends a request through Require. An empty username means "send no
// credentials at all".
func gated(t *testing.T, path, username, password string) *httptest.ResponseRecorder {
	t.Helper()
	reached := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("reached"))
	})
	handler := Require(Credentials{Username: testUser, Password: testPassword}, reached)

	req := httptest.NewRequest(http.MethodGet, path, nil)
	if username != "" {
		req.SetBasicAuth(username, password)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}
