package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"lessonHttp/internal/httpx"
	"net/http"
	"net/http/httptest"
	"testing"
)

type FakeTokenVerifier struct {
	gotUserID int
	isCall    bool
	verifyErr error
}

func (v *FakeTokenVerifier) Verify(token string) (int, error) {
	v.isCall = true
	return v.gotUserID, v.verifyErr
}

var ErrVerify = errors.New("some verify error")

func TestAuth(t *testing.T) {
	var tests = []struct {
		name         string
		headerKey    string
		headerFlag   string
		isVerifyCall bool
		token        string
		userID       int
		verifyErr    error
		wantStatus   int
		wantErr      string
	}{
		{
			name:       "Autorization empty",
			headerKey:  "",
			verifyErr:  ErrVerify,
			wantStatus: 401,
			wantErr:    "unautorized",
		},
		{
			name:       "Bearer empty",
			headerKey:  "Autorization",
			headerFlag: "",
			verifyErr:  ErrVerify,
			wantStatus: 401,
			wantErr:    "unautorized",
		},
		{
			name:       "Token empty",
			headerKey:  "Autorization",
			headerFlag: "Bearer ",
			token:      "",
			verifyErr:  ErrVerify,
			wantStatus: 401,
			wantErr:    "unautorized",
		},
		{
			name:         "Verify error",
			headerKey:    "Autorization",
			headerFlag:   "Bearer ",
			token:        "12345",
			isVerifyCall: true,
			verifyErr:    ErrVerify,
			wantStatus:   401,
			wantErr:      "unautorized",
		},
		{
			name:         "correct",
			headerKey:    "Autorization",
			headerFlag:   "Bearer ",
			token:        "12345",
			isVerifyCall: true,
			wantStatus:   200,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			verifier := &FakeTokenVerifier{
				gotUserID: 1,
				verifyErr: tt.verifyErr,
			}

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/user/1", nil)
			req.Header.Set(tt.headerKey, tt.headerFlag+tt.token)

			modifContext := context.Background()
			mux := http.NewServeMux()
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				modifContext = r.Context()
				httpx.WriteJSON(w, http.StatusOK, nil)
			})
			mux.Handle("/user/1", Auth(verifier)(handler))
			mux.ServeHTTP(rec, req)

			if verifier.isCall != tt.isVerifyCall {
				fmt.Println("--")
				t.Fatalf("verify is called %t; want %t", verifier.isCall, tt.isVerifyCall)
			}

			if rec.Code != tt.wantStatus {
				t.Fatalf("status code %d; want %d", rec.Code, tt.wantStatus)
			}

			var responseErr *httpx.ErrorResponse
			err := json.NewDecoder(rec.Body).Decode(&responseErr)
			if responseErr != nil {
				if err != nil {
					t.Fatalf("decode error %v", err)
				}
				if responseErr.Error != tt.wantErr {
					t.Errorf("got error %q; want %q", responseErr.Error, tt.wantErr)
				}
				return
			}

			userID, ok := UserIDFromContext(modifContext)
			if !ok {
				t.Error("got empty userID;want userID")
			}

			if userID != verifier.gotUserID {
				t.Fatalf("userID %d; want %d", userID, verifier.gotUserID)
			}
		})
	}
}
