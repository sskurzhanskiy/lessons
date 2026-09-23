package password

import (
	"errors"
	"strings"
	"testing"
)

func TestVerify(t *testing.T) {
	var tests = []struct {
		name        string
		password    string
		oldSting    string
		dummyString string
		wantResult  bool
		wantError   error
	}{
		{
			name:        "password is correct",
			password:    "correct-password",
			oldSting:    "",
			dummyString: "",
			wantResult:  true,
		},
		{
			name:        "password is incorrect",
			password:    "incorrect-password",
			oldSting:    "",
			dummyString: "",
			wantResult:  false,
		},
		{
			name:        "invalid PHC structure",
			password:    "correct-password",
			oldSting:    "$argon2id$v=19",
			dummyString: "$argon2id$v=19$broken",
			wantResult:  false,
			wantError:   ErrInvalidHash,
		},
		{
			name:        "Unsupported algorithm",
			password:    "correct-password",
			oldSting:    "argon2id",
			dummyString: "argon42id",
			wantResult:  false,
			wantError:   ErrUnsupportedAlgorithm,
		},
		{
			name:        "Unsupported version",
			password:    "correct-password",
			oldSting:    "v=19",
			dummyString: "v=22",
			wantResult:  false,
			wantError:   ErrUnsupportedVersion,
		},
		{
			name:        "Unsupported parameters",
			password:    "correct-password",
			oldSting:    "m=65536,t=3,p=1",
			dummyString: "m=32768,t=3,p=1",
			wantResult:  false,
			wantError:   ErrUnsupportedParameters,
		},
	}

	hash, err := Hash("correct-password")
	if err != nil {
		t.Fatalf("create hash %v", err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modified := hash
			if tt.oldSting != "" && tt.dummyString != "" {
				modified = strings.Replace(hash, tt.oldSting, tt.dummyString, 1)
			}
			result, err := Verify(tt.password, modified)

			if result != tt.wantResult {
				t.Errorf("result %t; want: %t", result, tt.wantResult)
			}

			if !errors.Is(err, tt.wantError) {
				t.Errorf("verify %v, want %v", err, tt.wantError)
			}
		})
	}
}

func TestTwoHashesOfSamePasswordAreDifferent(t *testing.T) {
	hash1, err := Hash("correct-password")
	if err != nil {
		t.Fatalf("create hash1 %v", err)
	}
	hash2, err := Hash("correct-password")
	if err != nil {
		t.Fatalf("create hash2 %v", err)
	}

	if hash1 == hash2 {
		t.Errorf("hash1 == hash2; want not equel")
	}

	for _, hash := range []string{hash1, hash2} {
		ok, err := Verify("correct-password", hash)
		if err != nil {
			t.Fatalf("Verify() unexpected error: %v", err)
		}
		if !ok {
			t.Fatal("Verify() rejected the correct password")
		}
	}
}
