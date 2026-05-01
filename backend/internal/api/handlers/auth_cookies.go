package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	cookieAccessToken  = "access_token"
	cookieRefreshToken = "refresh_token"
)

func sameSiteMode(s string) http.SameSite {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "strict":
		return http.SameSiteStrictMode
	case "lax":
		return http.SameSiteLaxMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}

func (h *AuthHandler) setAuthCookies(c *gin.Context, accessToken, refreshToken string) {
	secure := h.cfg.Cookie.Secure
	ss := sameSiteMode(h.cfg.Cookie.SameSite)
	if ss == http.SameSiteNoneMode {
		secure = true
	}

	accessMax := int(h.cfg.JWT.AccessTTL.Seconds())
	refreshMax := int(h.cfg.JWT.RefreshTTL.Seconds())

	setHTTPOnlyCookie(c, cookieAccessToken, accessToken, accessMax, secure, ss)
	setHTTPOnlyCookie(c, cookieRefreshToken, refreshToken, refreshMax, secure, ss)
}

func setHTTPOnlyCookie(c *gin.Context, name, value string, maxAge int, secure bool, sameSite http.SameSite) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	})
}

func (h *AuthHandler) clearAuthCookies(c *gin.Context) {
	secure := h.cfg.Cookie.Secure
	ss := sameSiteMode(h.cfg.Cookie.SameSite)
	if ss == http.SameSiteNoneMode {
		secure = true
	}
	expireCookie(c, cookieAccessToken, secure, ss)
	expireCookie(c, cookieRefreshToken, secure, ss)
}

func expireCookie(c *gin.Context, name string, secure bool, sameSite http.SameSite) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	})
}
