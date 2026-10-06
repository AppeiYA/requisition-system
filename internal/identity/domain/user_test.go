package domain

import (
	"strings"
	"testing"
	"time"
)

func Test_NewUser(t *testing.T) {
	email, err := NewEmail("john.doe@example.com")
	if err != nil {
		t.Fatalf("unexpected error creating email: %v", err)
	}

	passwordHash, err := NewPasswordHash("hashed_password_123")
	if err != nil {
		t.Fatalf("unexpected error creating password hash: %v", err)
	}

	tests := []struct {
		name        string
		firstName   string
		lastName    string
		email       *Email
		passHash    *PasswordHash
		wantErr     bool
		expectedErr error
	}{
		{
			name:      "successful user creation",
			firstName: "John",
			lastName:  "Doe",
			email:     email,
			passHash:  passwordHash,
			wantErr:   false,
		},
		{
			name:        "firstName is empty",
			firstName:   "",
			lastName:    "Doe",
			email:       email,
			passHash:    passwordHash,
			wantErr:     true,
			expectedErr: ErrUsernameTooShort(FirstName),
		},
		{
			name:        "firstName is too short",
			firstName:   "J",
			lastName:    "Doe",
			email:       email,
			passHash:    passwordHash,
			wantErr:     true,
			expectedErr: ErrUsernameTooShort(FirstName),
		},
		{
			name:        "firstName is too long",
			firstName:   strings.Repeat("a", 101),
			lastName:    "Doe",
			email:       email,
			passHash:    passwordHash,
			wantErr:     true,
			expectedErr: ErrUsernameTooLong(FirstName),
		},
		{
			name:        "lastName is empty",
			firstName:   "John",
			lastName:    "",
			email:       email,
			passHash:    passwordHash,
			wantErr:     true,
			expectedErr: ErrUsernameTooShort(LastName),
		},
		{
			name:        "lastName is too short",
			firstName:   "John",
			lastName:    "D",
			email:       email,
			passHash:    passwordHash,
			wantErr:     true,
			expectedErr: ErrUsernameTooShort(LastName),
		},
		{
			name:        "lastName is too long",
			firstName:   "John",
			lastName:    strings.Repeat("a", 101),
			email:       email,
			passHash:    passwordHash,
			wantErr:     true,
			expectedErr: ErrUsernameTooLong(LastName),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now()
			user, err := NewUser(tt.firstName, tt.lastName, tt.email, tt.passHash)
			after := time.Now()

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewUser() error = nil, wantErr %v", tt.wantErr)
				}
				if user != nil {
					t.Errorf("NewUser() = %v, want nil", user)
				}
				if tt.expectedErr != nil && err != nil {
					if err.Error() != tt.expectedErr.Error() {
						t.Errorf("NewUser() error = %v, want %v", err, tt.expectedErr)
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("NewUser() unexpected error = %v", err)
			}
			if user == nil {
				t.Fatalf("NewUser() returned nil user")
			}

			if user.FirstName().Value() != tt.firstName {
				t.Errorf("user.FirstName().Value() = %v, want %v", user.FirstName().Value(), tt.firstName)
			}
			if user.LastName().Value() != tt.lastName {
				t.Errorf("user.LastName().Value() = %v, want %v", user.LastName().Value(), tt.lastName)
			}
			if user.Email() != tt.email {
				t.Errorf("user.Email() = %v, want %v", user.Email(), tt.email)
			}
			if user.PasswordHash() != tt.passHash {
				t.Errorf("user.PasswordHash() = %v, want %v", user.PasswordHash(), tt.passHash)
			}
			if user.Status() != ACTIVE {
				t.Errorf("user.Status() = %v, want %v", user.Status(), ACTIVE)
			}
			if user.CreatedAt().Before(before) || user.CreatedAt().After(after) {
				t.Errorf("user.CreatedAt() = %v, expected between %v and %v", user.CreatedAt(), before, after)
			}
			if user.UpdatedAt() != nil {
				t.Errorf("user.UpdatedAt() = %v, want nil", user.UpdatedAt())
			}
		})
	}
}

func Test_ReconstituteUser(t *testing.T) {
	now := time.Now()
	past := now.Add(-24 * time.Hour)
	recently := now.Add(-1 * time.Hour)

	tests := []struct {
		name         string
		firstName    string
		lastName     string
		email        string
		passwordHash string
		status       string
		createdAt    time.Time
		updatedAt    *time.Time
		wantErr      bool
		expectedErr  error
	}{
		{
			name:         "successful reconstitution without updatedAt",
			firstName:    "John",
			lastName:     "Doe",
			email:        "john.doe@example.com",
			passwordHash: "hashed_password",
			status:       "active",
			createdAt:    past,
			updatedAt:    nil,
			wantErr:      false,
		},
		{
			name:         "successful reconstitution with updatedAt",
			firstName:    "Jane",
			lastName:     "Doe",
			email:        "jane.doe@example.com",
			passwordHash: "hashed_password",
			status:       "suspended",
			createdAt:    past,
			updatedAt:    &recently,
			wantErr:      false,
		},
		{
			name:         "invalid firstName format (too short)",
			firstName:    "J",
			lastName:     "Doe",
			email:        "john.doe@example.com",
			passwordHash: "hashed_password",
			status:       "active",
			createdAt:    past,
			updatedAt:    nil,
			wantErr:      true,
			expectedErr:  ErrUsernameTooShort(FirstName),
		},
		{
			name:         "invalid firstName format (too long)",
			firstName:    strings.Repeat("a", 101),
			lastName:     "Doe",
			email:        "john.doe@example.com",
			passwordHash: "hashed_password",
			status:       "active",
			createdAt:    past,
			updatedAt:    nil,
			wantErr:      true,
			expectedErr:  ErrUsernameTooLong(FirstName),
		},
		{
			name:         "invalid lastName format (too short)",
			firstName:    "John",
			lastName:     "D",
			email:        "john.doe@example.com",
			passwordHash: "hashed_password",
			status:       "active",
			createdAt:    past,
			updatedAt:    nil,
			wantErr:      true,
			expectedErr:  ErrUsernameTooShort(LastName),
		},
		{
			name:         "invalid lastName format (too long)",
			firstName:    "John",
			lastName:     strings.Repeat("a", 101),
			email:        "john.doe@example.com",
			passwordHash: "hashed_password",
			status:       "active",
			createdAt:    past,
			updatedAt:    nil,
			wantErr:      true,
			expectedErr:  ErrUsernameTooLong(LastName),
		},
		{
			name:         "invalid email format",
			firstName:    "John",
			lastName:     "Doe",
			email:        "invalid-email",
			passwordHash: "hashed_password",
			status:       "active",
			createdAt:    past,
			updatedAt:    nil,
			wantErr:      true,
			expectedErr:  ErrInvalidEmailFormat,
		},
		{
			name:         "empty email",
			firstName:    "John",
			lastName:     "Doe",
			email:        "",
			passwordHash: "hashed_password",
			status:       "active",
			createdAt:    past,
			updatedAt:    nil,
			wantErr:      true,
			expectedErr:  ErrInvalidEmailFormat,
		},
		{
			name:         "empty password hash",
			firstName:    "John",
			lastName:     "Doe",
			email:        "john.doe@example.com",
			passwordHash: "",
			status:       "active",
			createdAt:    past,
			updatedAt:    nil,
			wantErr:      true,
			expectedErr:  ErrInvalidPasswordHash,
		},
		{
			name:         "invalid status",
			firstName:    "John",
			lastName:     "Doe",
			email:        "john.doe@example.com",
			passwordHash: "hashed_password",
			status:       "pending",
			createdAt:    past,
			updatedAt:    nil,
			wantErr:      true,
			expectedErr:  ErrInvalidStatus,
		},
		{
			name:         "empty status",
			firstName:    "John",
			lastName:     "Doe",
			email:        "john.doe@example.com",
			passwordHash: "hashed_password",
			status:       "",
			createdAt:    past,
			updatedAt:    nil,
			wantErr:      true,
			expectedErr:  ErrInvalidStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := ReconstituteUser(
				tt.firstName,
				tt.lastName,
				tt.email,
				tt.passwordHash,
				tt.status,
				tt.createdAt,
				tt.updatedAt,
			)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ReconstituteUser() error = nil, wantErr %v", tt.wantErr)
				}
				if user != nil {
					t.Errorf("ReconstituteUser() = %v, want nil", user)
				}
				if tt.expectedErr != nil && err != nil {
					if err.Error() != tt.expectedErr.Error() {
						t.Errorf("ReconstituteUser() error = %v, want %v", err, tt.expectedErr)
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("ReconstituteUser() unexpected error = %v", err)
			}
			if user == nil {
				t.Fatalf("ReconstituteUser() returned nil user")
			}

			if user.FirstName().Value() != tt.firstName {
				t.Errorf("user.FirstName().Value() = %v, want %v", user.FirstName().Value(), tt.firstName)
			}
			if user.LastName().Value() != tt.lastName {
				t.Errorf("user.LastName().Value() = %v, want %v", user.LastName().Value(), tt.lastName)
			}
			if user.Email().Value() != tt.email {
				t.Errorf("user.Email().Value() = %v, want %v", user.Email().Value(), tt.email)
			}
			if user.PasswordHash().Value() != tt.passwordHash {
				t.Errorf("user.PasswordHash().Value() = %v, want %v", user.PasswordHash().Value(), tt.passwordHash)
			}
			if user.Status().Value() != tt.status {
				t.Errorf("user.Status().Value() = %v, want %v", user.Status().Value(), tt.status)
			}
			if !user.CreatedAt().Equal(tt.createdAt) {
				t.Errorf("user.CreatedAt() = %v, want %v", user.CreatedAt(), tt.createdAt)
			}
			if tt.updatedAt == nil {
				if user.UpdatedAt() != nil {
					t.Errorf("user.UpdatedAt() = %v, want nil", user.UpdatedAt())
				}
			} else {
				if user.UpdatedAt() == nil || !user.UpdatedAt().Equal(*tt.updatedAt) {
					t.Errorf("user.UpdatedAt() = %v, want %v", user.UpdatedAt(), tt.updatedAt)
				}
			}
		})
	}
}

func Test_User_UpdateStatus(t *testing.T) {
	tests := []struct {
		name          string
		initialStatus Status
		newStatus     Status
		wantErr       bool
		expectedErr   error
	}{
		{
			name:          "valid transition from ACTIVE to SUSPENDED",
			initialStatus: ACTIVE,
			newStatus:     SUSPENDED,
			wantErr:       false,
		},
		{
			name:          "valid transition from SUSPENDED to ACTIVE",
			initialStatus: SUSPENDED,
			newStatus:     ACTIVE,
			wantErr:       false,
		},
		{
			name:          "invalid transition from ACTIVE to ACTIVE",
			initialStatus: ACTIVE,
			newStatus:     ACTIVE,
			wantErr:       true,
			expectedErr:   ErrInvalidStatus,
		},
		{
			name:          "invalid transition from SUSPENDED to SUSPENDED",
			initialStatus: SUSPENDED,
			newStatus:     SUSPENDED,
			wantErr:       true,
			expectedErr:   ErrInvalidStatus,
		},
		{
			name:          "invalid transition to unknown status",
			initialStatus: ACTIVE,
			newStatus:     Status("unknown"),
			wantErr:       true,
			expectedErr:   ErrInvalidStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			email, _ := NewEmail("john.doe@example.com")
			pass, _ := NewPasswordHash("hashed_password")

			var user *User
			var err error
			if tt.initialStatus == ACTIVE {
				user, err = NewUser("John", "Doe", email, pass)
			} else {
				user, err = ReconstituteUser("John", "Doe", "john.doe@example.com", "hashed_password", string(tt.initialStatus), time.Now(), nil)
			}
			if err != nil {
				t.Fatalf("failed to create initial user: %v", err)
			}

			beforeUpdate := time.Now()
			updateErr := user.UpdateStatus(tt.newStatus)
			afterUpdate := time.Now()

			if tt.wantErr {
				if updateErr == nil {
					t.Errorf("user.UpdateStatus() error = nil, wantErr %v", tt.wantErr)
				}
				if tt.expectedErr != nil && updateErr != tt.expectedErr {
					t.Errorf("user.UpdateStatus() error = %v, want %v", updateErr, tt.expectedErr)
				}
				if user.Status() != tt.initialStatus {
					t.Errorf("user.Status() changed to %v, want initial status %v", user.Status(), tt.initialStatus)
				}
				if user.UpdatedAt() != nil {
					t.Errorf("user.UpdatedAt() = %v, want nil on failed update", user.UpdatedAt())
				}
				return
			}

			if updateErr != nil {
				t.Fatalf("user.UpdateStatus() unexpected error = %v", updateErr)
			}
			if user.Status() != tt.newStatus {
				t.Errorf("user.Status() = %v, want %v", user.Status(), tt.newStatus)
			}
			if user.UpdatedAt() == nil {
				t.Fatalf("user.UpdatedAt() is nil, expected updated timestamp")
			}
			if user.UpdatedAt().Before(beforeUpdate) || user.UpdatedAt().After(afterUpdate) {
				t.Errorf("user.UpdatedAt() = %v, expected between %v and %v", user.UpdatedAt(), beforeUpdate, afterUpdate)
			}
		})
	}
}