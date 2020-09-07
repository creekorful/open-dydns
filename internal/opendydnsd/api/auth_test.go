package api

import (
	"encoding/base64"
	"encoding/json"
	"github.com/creekorful/open-dydns/pkg/proto"
	"strings"
	"testing"
	"time"
)

func TestMakeToken(t *testing.T) {
	token := encodeToken(t, 42, 0)
	if token.UserID != 42 {
		t.Error("wrong user id")
	}
}

func encodeToken(t *testing.T, userID uint, ttl time.Duration) proto.UserContext {
	token, err := makeToken(proto.UserContext{UserID: userID}, []byte("test"), ttl)
	if err != nil {
		t.Error(err)
	}

	parts := strings.Split(token.Token, ".")
	if len(parts) != 3 {
		t.Errorf("Malformed JWT token")
	}

	payload := parts[1]
	bytes, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		t.Error(err)
	}

	var userCtx proto.UserContext
	if err := json.Unmarshal(bytes, &userCtx); err != nil {
		t.Error(err)
	}

	if userCtx.UserID != 42 {
		t.Error("wrong user id returned")
	}

	return userCtx
}

// TODO test token expiration
