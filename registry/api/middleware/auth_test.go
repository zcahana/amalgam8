// Copyright 2016 IBM Corporation
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"fmt"

	"github.com/amalgam8/amalgam8/registry/auth"
	"github.com/stretchr/testify/assert"
)

const (
	validToken   = "valid-token"
	invalidToken = "invalid-token"
)

type mockAuthenticator struct {
	authFunc func(token string) (*auth.Namespace, error)
}

func (ma *mockAuthenticator) Authenticate(token string) (*auth.Namespace, error) {
	if ma.authFunc != nil {
		return ma.authFunc(token)
	}
	return nil, fmt.Errorf("internal test error")
}

func TestEmptyTokenSuccess(t *testing.T) {
	ma := &mockAuthenticator{
		authFunc: func(token string) (*auth.Namespace, error) {
			if token == "" {
				namespace := auth.NamespaceFrom(validToken)
				return &namespace, nil
			}
			return nil, auth.ErrUnauthorized
		},
	}
	authMw := &AuthMiddleware{Authenticator: ma}

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://example.com/", nil)

	jrestServer(authMw, "/").ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
}

func TestEmptyTokenFailure(t *testing.T) {
	ma := &mockAuthenticator{
		authFunc: func(token string) (*auth.Namespace, error) {
			if token != "" {
				namespace := auth.NamespaceFrom(token)
				return &namespace, nil
			}
			return nil, auth.ErrUnauthorized
		},
	}
	authMw := &AuthMiddleware{Authenticator: ma}

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://example.com/", nil)

	jrestServer(authMw, "/").ServeHTTP(res, req)

	assert.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestTokenInvalid(t *testing.T) {
	ma := &mockAuthenticator{
		authFunc: func(token string) (*auth.Namespace, error) {
			if token == invalidToken {
				return nil, auth.ErrUnauthorized
			}
			namespace := auth.NamespaceFrom(token)
			return &namespace, nil
		},
	}
	authMw := &AuthMiddleware{Authenticator: ma}

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://example.com/", nil)
	req.Header.Set("Authorization", "Bearer "+invalidToken)

	jrestServer(authMw, "/").ServeHTTP(res, req)

	assert.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestDefaultAuthenticator(t *testing.T) {
	authMw := &AuthMiddleware{}

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://example.com/", nil)

	jrestServer(authMw, "/").ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
}

func TestCommunicationError(t *testing.T) {
	ma := &mockAuthenticator{
		authFunc: func(token string) (*auth.Namespace, error) {
			return nil, auth.ErrCommunicationError
		},
	}
	authMw := &AuthMiddleware{Authenticator: ma}

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "http://example.com/", nil)

	jrestServer(authMw, "/").ServeHTTP(res, req)

	assert.Equal(t, http.StatusServiceUnavailable, res.Code)
}
