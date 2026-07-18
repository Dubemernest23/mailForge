package auth

import (
	"strings"
	"testing"
)

func TestRegistrationRequest_val(t *testing.T) {
	// arrange, act, assert - all in here

	tests := []struct {
		name            string
		req             RegisterRequest
		wantErr         bool
		wantErrContains string
	}{
		{
			name: "valid request",
			req: RegisterRequest{
				Username: "duby_dev",
				Email:    "duby@example.com",
				Password: "Str0ng!pass",
			},
			wantErr: false,
		},
		{
			name: "missing_username",
			req: RegisterRequest{
				Username: "",
				Email:    "dubydev@mail.com",
				Password: "Str0ng!pass",
			},
			wantErr:         true,
			wantErrContains: "required",
		},
		{
			name: "username_too_short",
			req: RegisterRequest{
				Username: "ab",
				Email:    "dubydev@mail.com",
				Password: "Str0ng!pass",
			},
			wantErr:         true,
			wantErrContains: "at least",
		},
		{
			name: "username_with_disallowed_char",
			req: RegisterRequest{
				Username: "ab-uu",
				Email:    "dubydev@mail.com",
				Password: "Str0ng!pass",
			},
			wantErr:         true,
			wantErrContains: "alphanumeric",
		},
		{
			name: "malformed_email",
			req: RegisterRequest{
				Username: "dub_de",
				Email:    "dubydev-mail.com",
				Password: "Str0ng!pass",
			},
			wantErr:         true,
			wantErrContains: "valid email",
		},
		{
			name: "password_too_short",
			req: RegisterRequest{
				Username: "dub_dev",
				Email:    "dubydev@mail.com",
				Password: "St",
			},
			wantErr:         true,
			wantErrContains: "at least",
		},
		{
			name: "malformed_password",
			req: RegisterRequest{
				Username: "dubdev",
				Email:    "dubydev@mail.com",
				Password: "Strngpass",
			},
			wantErr:         true,
			wantErrContains: "special character",
		},
		{
			name: "missing_password",
			req: RegisterRequest{
				Username: "dubdev",
				Email:    "dubydev@mail.com",
				Password: "",
			},
			wantErr:         true,
			wantErrContains: "required",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.req.Validate()
			gotErr := err != nil

			if gotErr != tc.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tc.wantErr)
			}

			if tc.wantErr && !strings.Contains(err.Error(), tc.wantErrContains) {
				t.Errorf("Validate() error = %q, want it to contain %q", err.Error(), tc.wantErrContains)
			}

		})
	}

}

func TestLoginRequest_val(t *testing.T) {
	// almost the same as register except for the username cases
	// we will have 6 cases here - valid_req, required_email, required_password
	// malformed_email, malformed_password, password_too_short

	tests := []struct {
		name            string
		req             LoginRequest
		wantErr         bool
		wantErrContains string
	}{
		{
			name: "valid_req",
			req: LoginRequest{
				Email:    "duby@example.com",
				Password: "Str0ng!pass",
			},
			wantErr: false,
		},
		{
			name: "required_email",
			req: LoginRequest{
				Email:    "",
				Password: "Str0ng!pass",
			},
			wantErr:         true,
			wantErrContains: "required",
		},
		{
			name: "required_password",
			req: LoginRequest{
				Email:    "duby@example.com",
				Password: "",
			},
			wantErr:         true,
			wantErrContains: "required",
		},
		{
			name: "malformed_email",
			req: LoginRequest{
				Email:    "duby-example.com",
				Password: "Str0ng!pass",
			},
			wantErr:         true,
			wantErrContains: "valid email",
		},
		{
			name: "malformed_password",
			req: LoginRequest{
				Email:    "duby@example.com",
				Password: "Strngpass",
			},
			wantErr:         true,
			wantErrContains: "special character",
		},
		{
			name: "password_too_short",
			req: LoginRequest{
				Email:    "duby@example.com",
				Password: "Str",
			},
			wantErr:         true,
			wantErrContains: "at least",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.req.Validate()
			gotErr := err != nil

			if gotErr != tc.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tc.wantErr)
			}

			if tc.wantErr && !strings.Contains(err.Error(), tc.wantErrContains) {
				t.Errorf("Validate() error = %q, want it to contain %q", err.Error(), tc.wantErrContains)
			}
		})
	}
}

func TestRefreshRequest_val(t *testing.T) {
	// cases - valid-uuid, malformed_uuid, missing_uuid
	tests := []struct {
		name            string
		req             RefreshRequest
		wantErr         bool
		wantErrContains string
	}{
		{
			name: "valid_refreshtoken",
			req: RefreshRequest{
				RefreshToken: "3fa85f64-5717-4562-b3fc-2c963f66afa6",
			},
			wantErr: false,
		},
		{
			name: "malformed_refreshtoken",
			req: RefreshRequest{
				RefreshToken: "309ikrkkr",
			},
			wantErr:         true,
			wantErrContains: "valid UUIDv4",
		},
		{
			name: "missing_refreshtoken",
			req: RefreshRequest{
				RefreshToken: "",
			},
			wantErr:         true,
			wantErrContains: "required",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.req.Validate()
			gotErr := err != nil

			if gotErr != tc.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tc.wantErr)
			}

			if tc.wantErr && !strings.Contains(err.Error(), tc.wantErrContains) {
				t.Errorf("Validate() error = %q, want it to contain %q", err.Error(), tc.wantErrContains)
			}
		})
	}
}

func TestLogoutRequest_val(t *testing.T) {
	// same as TestRefreshToken

	tests := []struct {
		name            string
		req             LogoutRequest
		wantErr         bool
		wantErrContains string
	}{
		{
			name: "missing_token",
			req: LogoutRequest{
				RefreshToken: "",
			},
			wantErr:         true,
			wantErrContains: "required",
		},
		{
			name: "valid_token",
			req: LogoutRequest{
				RefreshToken: "3fa85f64-5717-4562-b3fc-2c963f66afa6",
			},
			wantErr: false,
		},
		{
			name: "malformed_token",
			req: LogoutRequest{
				RefreshToken: "3fa85f64a6",
			},
			wantErr:         true,
			wantErrContains: "valid UUIDv4",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.req.Validate()
			gotErr := err != nil

			if gotErr != tc.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tc.wantErr)
			}

			if tc.wantErr && !strings.Contains(err.Error(), tc.wantErrContains) {
				t.Errorf("Validate() error = %q, want it to contain %q", err.Error(), tc.wantErrContains)
			}
		})
	}
}
