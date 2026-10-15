package domain

import (
	"strings"
	"testing"
	"time"

	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
)

func Test_NewUser(t *testing.T) {
	t.Parallel()

	email, err := shareddomain.NewEmail("john.doe@example.com")
	if err != nil {
		t.Fatalf("unexpected error creating email: %v", err)
	}

	passwordHash, err := NewPasswordHash("hashed_password_123")
	if err != nil {
		t.Fatalf("unexpected error creating password hash: %v", err)
	}

	const (
		defaultID        = "usr_12345"
		defaultFirstName = "John"
		defaultLastName  = "Doe"
		defaultPhone     = "+14155551234"
		defaultStaffID   = "STF-1234"
	)

	tests := []struct {
		name        string
		id          string
		firstName   string
		lastName    string
		phone       string
		wantErr     bool
		expectedErr error
	}{
		{
			name:      "successful user creation",
			id:        defaultID,
			firstName: defaultFirstName,
			lastName:  defaultLastName,
			wantErr:   false,
		},
		{
			name:        "id is empty",
			id:          "",
			firstName:   defaultFirstName,
			lastName:    defaultLastName,
			wantErr:     true,
			expectedErr: ErrInvalidID,
		},
		{
			name:        "firstName is empty",
			id:          defaultID,
			firstName:   "",
			lastName:    defaultLastName,
			wantErr:     true,
			expectedErr: ErrUsernameTooShort(FirstName),
		},
		{
			name:        "firstName is too short",
			id:          defaultID,
			firstName:   "J",
			lastName:    defaultLastName,
			wantErr:     true,
			expectedErr: ErrUsernameTooShort(FirstName),
		},
		{
			name:        "firstName is too long",
			id:          defaultID,
			firstName:   strings.Repeat("a", 101),
			lastName:    defaultLastName,
			wantErr:     true,
			expectedErr: ErrUsernameTooLong(FirstName),
		},
		{
			name:        "lastName is empty",
			id:          defaultID,
			firstName:   defaultFirstName,
			lastName:    "",
			wantErr:     true,
			expectedErr: ErrUsernameTooShort(LastName),
		},
		{
			name:        "lastName is too short",
			id:          defaultID,
			firstName:   defaultFirstName,
			lastName:    "D",
			wantErr:     true,
			expectedErr: ErrUsernameTooShort(LastName),
		},
		{
			name:        "lastName is too long",
			id:          defaultID,
			firstName:   defaultFirstName,
			lastName:    strings.Repeat("a", 101),
			wantErr:     true,
			expectedErr: ErrUsernameTooLong(LastName),
		},
		{
			name:        "invalid phone number",
			id:          defaultID,
			firstName:   defaultFirstName,
			lastName:    defaultLastName,
			phone:       "not-a-phone",
			wantErr:     true,
			expectedErr: shareddomain.ErrInvalidPhoneNumber,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			phone := tt.phone
			if phone == "" {
				phone = defaultPhone
			}

			before := time.Now()
			user, err := NewUser(tt.id, tt.firstName, tt.lastName, phone, defaultStaffID, email, passwordHash)
			after := time.Now()

			if tt.wantErr {
				if err == nil {
					t.Fatalf("NewUser() error = nil, wantErr %v", tt.wantErr)
				}
				if user != nil {
					t.Errorf("NewUser() = %v, want nil", user)
				}
				if tt.expectedErr != nil && err.Error() != tt.expectedErr.Error() {
					t.Errorf("NewUser() error = %v, want %v", err, tt.expectedErr)
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
			if user.Phone().Value() != phone {
				t.Errorf("user.Phone().Value() = %v, want %v", user.Phone().Value(), phone)
			}
			if user.StaffID() != defaultStaffID {
				t.Errorf("user.StaffID() = %v, want %v", user.StaffID(), defaultStaffID)
			}
			if user.Email() != email {
				t.Errorf("user.Email() = %v, want %v", user.Email(), email)
			}
			if user.PasswordHash() != passwordHash {
				t.Errorf("user.PasswordHash() = %v, want %v", user.PasswordHash(), passwordHash)
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
	t.Parallel()

	now := time.Now()
	past := now.Add(-24 * time.Hour)
	recently := now.Add(-1 * time.Hour)

	type reconstituteInput struct {
		id           string
		firstName    string
		lastName     string
		phone        string
		staffID      string
		email        string
		passwordHash string
		status       string
		createdAt    time.Time
		updatedAt    *time.Time
	}

	validInput := func() reconstituteInput {
		return reconstituteInput{
			id:           "usr_12345",
			firstName:    "John",
			lastName:     "Doe",
			phone:        "+14155551234",
			staffID:      "STF-1234",
			email:        "john.doe@example.com",
			passwordHash: "hashed_password",
			status:       "active",
			createdAt:    past,
			updatedAt:    nil,
		}
	}

	tests := []struct {
		name        string
		mutate      func(*reconstituteInput)
		wantErr     bool
		expectedErr error
	}{
		{
			name:    "successful reconstitution without updatedAt",
			mutate:  nil,
			wantErr: false,
		},
		{
			name: "successful reconstitution with updatedAt",
			mutate: func(in *reconstituteInput) {
				in.firstName = "Jane"
				in.status = "suspended"
				in.updatedAt = &recently
			},
			wantErr: false,
		},
		{
			name: "invalid firstName format (too short)",
			mutate: func(in *reconstituteInput) {
				in.firstName = "J"
			},
			wantErr:     true,
			expectedErr: ErrUsernameTooShort(FirstName),
		},
		{
			name: "invalid firstName format (too long)",
			mutate: func(in *reconstituteInput) {
				in.firstName = strings.Repeat("a", 101)
			},
			wantErr:     true,
			expectedErr: ErrUsernameTooLong(FirstName),
		},
		{
			name: "invalid lastName format (too short)",
			mutate: func(in *reconstituteInput) {
				in.lastName = "D"
			},
			wantErr:     true,
			expectedErr: ErrUsernameTooShort(LastName),
		},
		{
			name: "invalid lastName format (too long)",
			mutate: func(in *reconstituteInput) {
				in.lastName = strings.Repeat("a", 101)
			},
			wantErr:     true,
			expectedErr: ErrUsernameTooLong(LastName),
		},
		{
			name: "invalid email format",
			mutate: func(in *reconstituteInput) {
				in.email = "invalid-email"
			},
			wantErr:     true,
			expectedErr: shareddomain.ErrInvalidEmailFormat,
		},
		{
			name: "invalid phone number",
			mutate: func(in *reconstituteInput) {
				in.phone = "not-a-phone"
			},
			wantErr:     true,
			expectedErr: shareddomain.ErrInvalidPhoneNumber,
		},
		{
			name: "empty email",
			mutate: func(in *reconstituteInput) {
				in.email = ""
			},
			wantErr:     true,
			expectedErr: shareddomain.ErrInvalidEmailFormat,
		},
		{
			name: "empty password hash",
			mutate: func(in *reconstituteInput) {
				in.passwordHash = ""
			},
			wantErr:     true,
			expectedErr: ErrInvalidPasswordHash,
		},
		{
			name: "invalid status",
			mutate: func(in *reconstituteInput) {
				in.status = "pending"
			},
			wantErr:     true,
			expectedErr: ErrInvalidStatus,
		},
		{
			name: "empty status",
			mutate: func(in *reconstituteInput) {
				in.status = ""
			},
			wantErr:     true,
			expectedErr: ErrInvalidStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			in := validInput()
			if tt.mutate != nil {
				tt.mutate(&in)
			}

			user, err := ReconstituteUser(
				in.id,
				in.firstName,
				in.lastName,
				in.phone,
				in.staffID,
				in.email,
				in.passwordHash,
				in.status,
				in.createdAt,
				in.updatedAt,
			)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("ReconstituteUser() error = nil, wantErr %v", tt.wantErr)
				}
				if user != nil {
					t.Errorf("ReconstituteUser() = %v, want nil", user)
				}
				if tt.expectedErr != nil && err.Error() != tt.expectedErr.Error() {
					t.Errorf("ReconstituteUser() error = %v, want %v", err, tt.expectedErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("ReconstituteUser() unexpected error = %v", err)
			}
			if user == nil {
				t.Fatalf("ReconstituteUser() returned nil user")
			}

			if user.FirstName().Value() != in.firstName {
				t.Errorf("user.FirstName().Value() = %v, want %v", user.FirstName().Value(), in.firstName)
			}
			if user.LastName().Value() != in.lastName {
				t.Errorf("user.LastName().Value() = %v, want %v", user.LastName().Value(), in.lastName)
			}
			if user.Phone().Value() != in.phone {
				t.Errorf("user.Phone().Value() = %v, want %v", user.Phone().Value(), in.phone)
			}
			if user.StaffID() != in.staffID {
				t.Errorf("user.StaffID() = %v, want %v", user.StaffID(), in.staffID)
			}
			if user.Email().Value() != in.email {
				t.Errorf("user.Email().Value() = %v, want %v", user.Email().Value(), in.email)
			}
			if user.PasswordHash().Value() != in.passwordHash {
				t.Errorf("user.PasswordHash().Value() = %v, want %v", user.PasswordHash().Value(), in.passwordHash)
			}
			if user.Status().Value() != in.status {
				t.Errorf("user.Status().Value() = %v, want %v", user.Status().Value(), in.status)
			}
			if !user.CreatedAt().Equal(in.createdAt) {
				t.Errorf("user.CreatedAt() = %v, want %v", user.CreatedAt(), in.createdAt)
			}
			if in.updatedAt == nil {
				if user.UpdatedAt() != nil {
					t.Errorf("user.UpdatedAt() = %v, want nil", user.UpdatedAt())
				}
			} else {
				if user.UpdatedAt() == nil || !user.UpdatedAt().Equal(*in.updatedAt) {
					t.Errorf("user.UpdatedAt() = %v, want %v", user.UpdatedAt(), in.updatedAt)
				}
			}
		})
	}
}

func Test_User_UpdateStatus(t *testing.T) {
	t.Parallel()

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
			t.Parallel()

			user, err := ReconstituteUser(
				"usr_12345",
				"John",
				"Doe",
				"+14155551234",
				"STF-1234",
				"john.doe@example.com",
				"hashed_password",
				string(tt.initialStatus),
				time.Now(),
				nil,
			)
			if err != nil {
				t.Fatalf("failed to create initial user: %v", err)
			}

			beforeUpdate := time.Now()
			updateErr := user.UpdateStatus(tt.newStatus)
			afterUpdate := time.Now()

			if tt.wantErr {
				if updateErr == nil {
					t.Fatalf("user.UpdateStatus() error = nil, wantErr %v", tt.wantErr)
				}
				if tt.expectedErr != nil && updateErr.Error() != tt.expectedErr.Error() {
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
