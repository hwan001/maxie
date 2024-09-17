package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"fileoptimizer/common"
)

func TestAuthenticate(t *testing.T) {
	// 테스트를 위한 모의 서버 생성
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"token": "mocked-jwt-token"}`))
	}))
	defer mockServer.Close()

	// 테스트용 인증 정보
	creds := common.Credentials{
		Username: "testuser",
		Password: "testpassword",
	}

	// Authenticate 함수 호출
	token, err := Authenticate(mockServer.URL, creds)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// 결과 확인
	expectedToken := "mocked-jwt-token"
	if token != expectedToken {
		t.Fatalf("Expected token to be %v, got %v", expectedToken, token)
	}
}
