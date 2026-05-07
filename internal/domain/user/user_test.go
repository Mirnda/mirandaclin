package user

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	var password = "mirandaClin"
	salt1, hash1, err := hashPassword(password)
	if err != nil {
		t.Fatalf("hashPassword erro: %v", err)
	}

	salt2, hash2, err := hashPassword(password)
	if err != nil {
		t.Fatalf("hashPassword erro: %v", err)
	}

	if salt1 == salt2 {
		t.Error("salt deve ser diferente a cada chamada")
	}
	if hash1 == hash2 {
		t.Error("hash deve ser diferente a cada chamada")
	}
}

func TestGenerateToken(t *testing.T) {
	type Params struct {
		onlyNumbers bool
		length      int64
	}
	tests := []Params{
		{onlyNumbers: true, length: 6},
		{onlyNumbers: true, length: 6},
		{onlyNumbers: true, length: 6},
		{onlyNumbers: true, length: 6},
		{onlyNumbers: true, length: 6},
		{onlyNumbers: false, length: 6},
		{onlyNumbers: false, length: 6},
		{onlyNumbers: false, length: 6},
		{onlyNumbers: false, length: 6},
		{onlyNumbers: true, length: 30},
		{onlyNumbers: false, length: 20},
		{onlyNumbers: false, length: 20},
		{onlyNumbers: false, length: 20},
	}

	for i, test := range tests {
		token, err := generateToken(test.length, test.onlyNumbers)
		if err != nil {
			t.Error(err)
		}

		t.Logf("t%d: %s", i, token)
	}
}
