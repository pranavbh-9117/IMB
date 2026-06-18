package password_test

import (
	"testing"

	"github.com/pranavbh-9117/IMB/pkg/password"
)

func TestGenerateTemp(t *testing.T) {
	length := 12
	temp1, err := password.GenerateTemp(length)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(temp1) != length {
		t.Errorf("expected length %d, got %d", length, len(temp1))
	}

	temp2, err := password.GenerateTemp(length)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if temp1 == temp2 {
		t.Errorf("expected random passwords to be unique, got two identical: %s", temp1)
	}
}
