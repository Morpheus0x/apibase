package helper

import "net/http"

// overwrite existing request cookie or adds it if missing
func OverwriteRequestCookie(request *http.Request, newCookie *http.Cookie) {
	newCookieHeader := ""
	existingCookieOverwritten := false
	requestCookies := request.Cookies()
	for _, cookie := range requestCookies {
		if newCookieHeader != "" {
			newCookieHeader += "; "
		}
		if cookie.Name == newCookie.Name {
			existingCookieOverwritten = true
			// Strip additional response cookie values to produce valid request cookie
			newCookieHeader += (&http.Cookie{Name: newCookie.Name, Value: newCookie.Value}).String()
		} else {
			newCookieHeader += cookie.String()
		}
	}
	if !existingCookieOverwritten {
		if newCookieHeader != "" {
			newCookieHeader += "; "
		}
		// Strip additional response cookie values to produce valid request cookie
		newCookieHeader += (&http.Cookie{Name: newCookie.Name, Value: newCookie.Value}).String()
	}
	request.Header.Set("Cookie", newCookieHeader)
}
