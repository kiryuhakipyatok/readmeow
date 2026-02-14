package utils

import (
	"bytes"
	"errors"
	"mime/multipart"
	"net/http/httptest"
	"reflect"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

func GenerateJWT(tokenTTL time.Duration, id, secret string) (string, *time.Time, error) {
	now := time.Now()
	t := now.Add(tokenTTL)
	ttl := jwt.NewNumericDate(t)
	iat := jwt.NewNumericDate(now)
	jti := uuid.New().String()
	claims := jwt.RegisteredClaims{
		Subject:   id,
		ExpiresAt: ttl,
		IssuedAt:  iat,
		ID:        jti,
		Issuer:    "readmeow",
		Audience:  []string{"readmeow-users"},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", nil, err
	}
	return jwtToken, &t, nil
}

func ToMap[T any](req T, typee string) (any, error) {
	sort := map[string]string{}
	filter := map[string][]string{}
	v := reflect.ValueOf(req)
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		fV := v.Field(i)
		fN := t.Field(i).Name
		if !fV.IsZero() {
			switch typee {
			case "sort":
				sort[fN] = fV.Interface().(string)
			case "filter":
				filter[fN] = fV.Interface().([]string)
			default:
				return nil, errors.New("invalid map type")
			}
		}
	}
	switch typee {
	case "sort":
		return sort, nil
	case "filter":
		return filter, nil
	default:
		return nil, errors.New("invalid map type")
	}
}

func StrToPtr(v string) *string {
	return &v
}

func NewTestFileHeader() *multipart.FileHeader {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, _ := writer.CreateFormFile("avatar", "test.png")
	part.Write([]byte("fake image content"))
	writer.Close()

	req := httptest.NewRequest("POST", "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.ParseMultipartForm(10 << 20)

	_, fh, _ := req.FormFile("avatar")
	return fh
}
