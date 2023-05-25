package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin/plugintest"
)

func TestServeHTTP(t *testing.T) {
	for name, tc := range map[string]struct {
		route          string
		expectedStatus int
		expectedError  string
	}{
		"blank path": {
			route:          "/",
			expectedStatus: 400,
			expectedError:  "no handler for route /",
		},
		"invalid path": {
			route:          "/invalid",
			expectedStatus: 400,
			expectedError:  "no handler for route /invalid",
		},
		"valid path": {
			route:          "/remove_user_from_all_teams_and_channels",
			expectedStatus: 405,
			expectedError:  "unexpected HTTP method GET. Should be POST",
		},
	} {
		t.Run(name, func(t *testing.T) {
			plugin := Plugin{}
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, tc.route, nil)

			plugin.ServeHTTP(nil, w, r)

			result := w.Result()
			require.NotNil(t, result)
			defer result.Body.Close()
			bodyBytes, err := io.ReadAll(result.Body)
			require.NoError(t, err)

			require.Equal(t, result.Header.Get("Content-Type"), "application/json")
			require.Equal(t, tc.expectedStatus, result.StatusCode)

			if tc.expectedError != "" {
				var errResponse ErrorResponse
				err := json.Unmarshal(bodyBytes, &errResponse)
				require.NoError(t, err)
				require.Equal(t, tc.expectedError, errResponse.Error)
			}
		})
	}
}

func TestHandleRemoveUserFromAllTeamsAndChannels(t *testing.T) {
	for name, tc := range map[string]struct {
		makeRequest    func(api *plugintest.API) *http.Request
		expectedStatus int
		expectedError  string
	}{
		"invalid http method": {
			makeRequest: func(_ *plugintest.API) *http.Request {
				r := httptest.NewRequest(http.MethodGet, "/remove_user_from_all_teams_and_channels", nil)
				return r
			},
			expectedStatus: 405,
			expectedError:  "unexpected HTTP method GET. Should be POST",
		},
		"missing user session": {
			makeRequest: func(_ *plugintest.API) *http.Request {
				r := httptest.NewRequest(http.MethodPost, "/remove_user_from_all_teams_and_channels", nil)
				return r
			},
			expectedStatus: 401,
			expectedError:  "request is not from an authenticated user",
		},
		"user not found": {
			makeRequest: func(api *plugintest.API) *http.Request {
				r := httptest.NewRequest(http.MethodPost, "/remove_user_from_all_teams_and_channels", nil)
				r.Header.Set("Mattermost-User-Id", "someuserid")

				api.On("GetUser", "someuserid").Return(nil, model.NewAppError("", "", nil, "user not found", 404))

				return r
			},
			expectedStatus: 401,
			expectedError:  "error verifying whether user someuserid is a system admin: failed to get user with id someuserid: : , user not found",
		},
		"user is not sysadmin": {
			makeRequest: func(api *plugintest.API) *http.Request {
				r := httptest.NewRequest(http.MethodPost, "/remove_user_from_all_teams_and_channels", nil)
				r.Header.Set("Mattermost-User-Id", "someuserid")

				api.On("GetUser", "someuserid").Return(&model.User{
					Roles: "system_user",
				}, nil)

				return r
			},
			expectedStatus: 401,
			expectedError:  "error verifying whether user someuserid is a system admin: user is not a system admin",
		},
		"missing user info in request": {
			makeRequest: func(api *plugintest.API) *http.Request {
				r := httptest.NewRequest(http.MethodPost, "/remove_user_from_all_teams_and_channels", nil)
				r.Header.Set("Mattermost-User-Id", "someuserid")

				api.On("GetUser", "someuserid").Return(&model.User{
					Roles: "system_user system_admin",
				}, nil)

				return r
			},
			expectedStatus: 500,
			expectedError:  "error processing request: error decoding user info payload: EOF",
		},
	} {
		t.Run(name, func(t *testing.T) {
			p := &Plugin{}
			api := &plugintest.API{}
			p.SetAPI(api)

			w := httptest.NewRecorder()
			r := tc.makeRequest(api)

			p.ServeHTTP(nil, w, r)

			result := w.Result()
			require.NotNil(t, result)
			defer result.Body.Close()
			bodyBytes, err := io.ReadAll(result.Body)
			require.NoError(t, err)

			require.Equal(t, result.Header.Get("Content-Type"), "application/json")
			require.Equal(t, tc.expectedStatus, result.StatusCode)

			if tc.expectedError != "" {
				var errResponse ErrorResponse
				err := json.Unmarshal(bodyBytes, &errResponse)
				require.NoError(t, err)
				require.Equal(t, tc.expectedError, errResponse.Error)
			}
		})
	}
}
