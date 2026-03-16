package auth

import "testing"

func TestHashPassword(t *testing.T) {
	password := ""
	_, err := HashPassword(password)

	if err.Error() != "password too short" {
		t.Fatal(err)
	}
	
}

func TestHashPassword_NotEmpty(t *testing.T) {
	password := "Abcd1234"

	got, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	if got == "" {
		t.Fatal("HashPassword() returned empty hash")
	}
	if got == password {
		t.Fatal("HashPassword() returned password in clear view")
	}
	if !ComparePasswords(got, password) {
		t.Fatal("hashed password does not match original password")
	}
}

func TestComparePasswords_Valid(t *testing.T) {
	password := "Abcd1234"
	hashPassword, err := HashPassword("Abcd1234")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	got := ComparePasswords(hashPassword, password)

	if got != true {
		t.Fatal("ComparePasswords() returned wrong result")
	}
}

func TestComparePasswords_Invalid(t *testing.T) {
	password := "Abcd12345"
	hashPassword, err := HashPassword("Abcd1234")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	got := ComparePasswords(hashPassword, password)

	if got == true {
		t.Fatal("ComparePasswords() returned true result")
	}
}
