package totp

import (
	"embed"
	"html/template"
	"net/http"
	"net/url"
)

const verifyQueryParam = "__torii_totp"
const verifyQueryValue = "verify"

//go:embed totp.html
var pageFS embed.FS

var challengeTemplate = template.Must(template.ParseFS(pageFS, "totp.html"))

type ChallengePageData struct {
	Action string
	Error  string
	Digits int
}

func IsVerifyRequest(r *http.Request) bool {
	return r.Method == http.MethodPost && r.URL.Query().Get(verifyQueryParam) == verifyQueryValue
}

func RenderChallenge(w http.ResponseWriter, r *http.Request, digits int, errorMessage string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusUnauthorized)
	_ = challengeTemplate.Execute(w, ChallengePageData{
		Action: verifyAction(r.URL),
		Error:  errorMessage,
		Digits: digits,
	})
}

func RedirectAfterVerification(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, cleanVerificationURL(r.URL), http.StatusSeeOther)
}

func verifyAction(u *url.URL) string {
	next := copyURL(u)
	q := next.Query()
	q.Set(verifyQueryParam, verifyQueryValue)
	next.RawQuery = q.Encode()
	return next.RequestURI()
}

func cleanVerificationURL(u *url.URL) string {
	next := copyURL(u)
	q := next.Query()
	q.Del(verifyQueryParam)
	next.RawQuery = q.Encode()
	return next.RequestURI()
}

func copyURL(u *url.URL) *url.URL {
	next := *u
	return &next
}
