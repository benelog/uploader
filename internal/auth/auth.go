// Package auth resolves the HTTP Basic credentials the uploader requires and
// provides the gate that enforces them.
//
// The policy is: authentication is on by default. Unless they are given
// explicitly, the username defaults to the OS login account and the password
// is a freshly generated random string, printed once at startup.
package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"log"
	"net/http"
	"os"
	"os/user"
	"strings"
)

// fallbackUsername is used when the OS login account cannot be determined.
const fallbackUsername = "admin"

// passwordBytes is the entropy of a generated password (128 bits).
const passwordBytes = 16

// realm is the HTTP Basic realm shown in the browser's login prompt.
const realm = "uploader"

// Credentials are the username/password pair the gate checks against.
type Credentials struct {
	Username string
	Password string
	// DefaultedUsername reports whether Username came from the OS login
	// account rather than from --user.
	DefaultedUsername bool
	// GeneratedPassword reports whether Password was randomly generated
	// rather than supplied by the operator. Only a generated password is safe
	// to echo to the console.
	GeneratedPassword bool
}

// Resolve builds the credentials from the (possibly empty) command line
// values, filling in the OS login name and a random password as needed.
func Resolve(username, password string) (Credentials, error) {
	c := Credentials{
		Username: strings.TrimSpace(username),
		Password: password,
	}
	if c.Username == "" {
		c.Username = osUsername()
		c.DefaultedUsername = true
	}
	if c.Password == "" {
		generated, err := generatePassword()
		if err != nil {
			return Credentials{}, err
		}
		c.Password = generated
		c.GeneratedPassword = true
	}
	return c, nil
}

// generatePassword returns a random, URL-safe password.
func generatePassword() (string, error) {
	buf := make([]byte, passwordBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// digests hashes the expected credentials, so the per-request gate compares
// against precomputed values instead of re-hashing them on every request.
func (c Credentials) digests() (user, pass [sha256.Size]byte) {
	return sha256.Sum256([]byte(c.Username)), sha256.Sum256([]byte(c.Password))
}

// matches compares the presented credentials against the expected digests in
// constant time. Hashing the presented values first keeps the comparison
// length-independent, so neither value's length leaks.
func matches(wantUser, wantPass [sha256.Size]byte, username, password string) bool {
	gotUser, gotPass := sha256.Sum256([]byte(username)), sha256.Sum256([]byte(password))
	userOK := subtle.ConstantTimeCompare(gotUser[:], wantUser[:]) == 1
	passOK := subtle.ConstantTimeCompare(gotPass[:], wantPass[:]) == 1
	return userOK && passOK
}

// Require wraps a handler so that every request below it, whatever the route,
// has to carry the expected Basic credentials. Wrapping the whole mux rather
// than each handler is what keeps a newly added route protected by default.
func Require(creds Credentials, next http.Handler) http.Handler {
	wantUser, wantPass := creds.digests()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok || !matches(wantUser, wantPass, username, password) {
			log.Printf("unauthorized request from %s: %s %s", r.RemoteAddr, r.Method, r.URL.Path)
			// The realm header is what makes the browser show a login prompt.
			w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`", charset="UTF-8"`)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// osUsername returns the login account of the current process owner, stripping
// any Windows "DOMAIN\" prefix, and falls back to the environment and finally
// to a fixed name.
func osUsername() string {
	if u, err := user.Current(); err == nil && u.Username != "" {
		name := u.Username
		if i := strings.LastIndexAny(name, `\/`); i >= 0 {
			name = name[i+1:]
		}
		if name != "" {
			return name
		}
	}
	for _, key := range []string{"USER", "USERNAME", "LOGNAME"} {
		if v := strings.TrimSpace(os.Getenv(key)); v != "" {
			return v
		}
	}
	return fallbackUsername
}
