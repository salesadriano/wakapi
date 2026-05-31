package utils

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/muety/wakapi/config"
)

func SetWebAuthnSession(session *webauthn.SessionData, r *http.Request, w http.ResponseWriter) error {
	sess, _ := config.GetSessionStore().Get(r, config.CookieKeySession)
	sess.Values[config.SessionValueWebAuthn] = session

	if config.Get().Security.CookieMaxAgeSec > 0 {
		sess.Values[config.SessionValueWebAuthnExpiresAt] = time.Now().Add(time.Duration(config.Get().Security.CookieMaxAgeSec) * time.Second).Unix()
	}

	if sess.Options != nil {
		sess.Options.MaxAge = config.Get().Security.CookieMaxAgeSec
	}

	return sess.Save(r, w)
}

func GetWebAuthnSession(r *http.Request) (*webauthn.SessionData, error) {
	sess, _ := config.GetSessionStore().Get(r, config.CookieKeySession)
	val, ok := sess.Values[config.SessionValueWebAuthn]
	if !ok {
		return nil, fmt.Errorf("webauthn session missing")
	}
	session, ok := val.(*webauthn.SessionData)
	if !ok {
		config.Log().Error("webauthn session data has invalid type")
		return nil, fmt.Errorf("webauthn session data has invalid type")
	}

	expiresAtRaw, hasExpiresAt := sess.Values[config.SessionValueWebAuthnExpiresAt]
	if config.Get().Security.CookieMaxAgeSec > 0 && !hasExpiresAt {
		return nil, fmt.Errorf("webauthn session expiry missing")
	}

	if hasExpiresAt {
		var expiresAt int64
		switch value := expiresAtRaw.(type) {
		case int64:
			expiresAt = value
		case int:
			expiresAt = int64(value)
		case int32:
			expiresAt = int64(value)
		case uint:
			expiresAt = int64(value)
		case uint32:
			expiresAt = int64(value)
		case uint64:
			expiresAt = int64(value)
		case float64:
			expiresAt = int64(value)
		default:
			return nil, fmt.Errorf("webauthn session expiry has invalid type")
		}

		if time.Now().Unix() > expiresAt {
			return nil, fmt.Errorf("webauthn session expired")
		}
	}

	return session, nil
}
