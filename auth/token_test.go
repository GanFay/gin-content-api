package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var userID = 5

func TestAccessSecret(t *testing.T) {
	if GetAccessSecret() == nil {
		t.Fatal("getAccessSecret() returns nil.")
	}
}
func TestRefreshSecret(t *testing.T) {
	if GetRefreshSecret() == nil {
		t.Fatal("GetRefreshSecret() returns nil.")
	}
}

func TestGenerateAccessJWT(t *testing.T) {

	got, err := GenerateAccessJWT(userID)
	if err != nil {
		t.Fatal("message: ", err)
	}

	if got == "" {
		t.Fatal("got empty string")
	}

	parseID, err := ParseJWTAccess(got)
	if err != nil {
		t.Fatal("message: ", err)
	}
	if userID != parseID {
		t.Fatal("got: ", parseID, "want: ", userID)
	}
	t.Log(got)
}

func TestGenerateRefreshJWT(t *testing.T) {

	got, err := GenerateRefreshJWT(userID)
	if err != nil {
		t.Fatal("message: ", err)
	}

	if got == "" {
		t.Fatal("got empty string")
	}

	parseID, err := ParseJWTRefresh(got)
	if err != nil {
		t.Fatal("message: ", err)
	}
	if userID != parseID {
		t.Fatal("got: ", parseID, "want: ", userID)
	}
	t.Log(got)
}

func TestParseAccessJWT_Valid(t *testing.T) {
	jwt, err := GenerateAccessJWT(userID)
	if err != nil {
		t.Fatal("message: ", err)
	}
	got, err := ParseJWTAccess(jwt)
	if err != nil {
		t.Fatal("message: ", err)
	}
	if got != userID {
		t.Fatal("got: ", got, "want: ", userID)
	}
}

func TestParseRefreshJWT_Valid(t *testing.T) {
	jwt, err := GenerateRefreshJWT(userID)
	if err != nil {
		t.Fatal("message: ", err)
	}
	got, err := ParseJWTRefresh(jwt)
	if err != nil {
		t.Fatal("message: ", err)
	}
	if got != userID {
		t.Fatal("got: ", got, "want: ", userID)
	}
}

func TestParseRefreshJWT_Invalid(t *testing.T) {
	_, err := ParseJWTRefresh("invalid_token")
	if err == nil {
		t.Fatal("message: err can't be nil (refresh)")
	}
}
func TestParseAccessJWT_Invalid(t *testing.T) {
	_, err := ParseJWTAccess("invalid_token")
	if err == nil {
		t.Fatal("message: err can't be nil (access)")
	}
}

func TestParseJWT_EmptyString(t *testing.T) {
	_, err := ParseJWTAccess("")
	if err == nil {
		t.Fatal("message: err can't be nil (Access)")
	}
	_, err = ParseJWTRefresh("")
	if err == nil {
		t.Fatal("message: err can't be nil (Refresh)")
	}
}
func TestParseJWT_WrongSecret(t *testing.T) {
	jwt, err := GenerateRefreshJWT(userID)
	if err != nil {
		t.Fatal("message: ", err)
	}
	_, err = ParseJWTAccess(jwt)
	if err == nil {
		t.Fatal("message: error in JWT (another Secret match! Generate:Refresh, Parse: Access)")
	}
	jwt, err = GenerateAccessJWT(userID)
	if err != nil {
		t.Fatal("message: ", err)
	}
	_, err = ParseJWTRefresh(jwt)
	if err == nil {
		t.Fatal("message: error in JWT (another Secret match! Generate: Access, Parse:Refresh)")
	}
}

func TestParseJWT_MissingUserID(t *testing.T) {
	jwtFunc := func() (string, error) {
		claims := jwt.MapClaims{
			"exp": time.Now().Add(15 * time.Minute).Unix(),
			"iat": time.Now().Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signedToken, err := token.SignedString(GetAccessSecret())
		if err != nil {
			return "", err
		}
		return signedToken, nil
	}
	jwtWithOutUserID, err := jwtFunc()
	if err != nil {
		t.Fatal("message: ", err)
	}
	_, err = ParseJWTAccess(jwtWithOutUserID)
	if err == nil {
		t.Fatal("message: err can't be nil (Access)")
	}
	_, err = ParseJWTRefresh(jwtWithOutUserID)
	if err == nil {
		t.Fatal("message: err can't be nil (Refresh)")
	}
}

func TestParseJWT_MissingUserIDType(t *testing.T) {
	id := "5"
	jwtFunc := func(userID string) (string, error) {
		claims := jwt.MapClaims{
			"user_id": userID,
			"exp":     time.Now().Add(15 * time.Minute).Unix(),
			"iat":     time.Now().Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signedToken, err := token.SignedString(GetAccessSecret())
		if err != nil {
			return "", err
		}
		return signedToken, nil
	}
	jwtWithOutUserID, err := jwtFunc(id)
	if err != nil {
		t.Fatal("message: ", err)
	}
	_, err = ParseJWTAccess(jwtWithOutUserID)
	if err == nil {
		t.Fatal("want: invalid user_id type(access)")
	}
	_, err = ParseJWTRefresh(jwtWithOutUserID)
	if err == nil {
		t.Fatal("want: invalid user_id type(refresh)")
	}
}
