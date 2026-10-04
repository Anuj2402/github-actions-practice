package main

import (
	"os"
	"testing"
)

func TestGetEnv_ReturnsEnvValueWhenSet(t *testing.T) {
	os.Setenv("TEST_KEY", "custom_value")
	defer os.Unsetenv("TEST_KEY")

	got := getEnv("TEST_KEY", "fallback_value")
	want := "custom_value"

	if got != want {
		t.Errorf("getEnv() = %q, want %q", got, want)
	}
}

func TestGetEnv_ReturnsFallbackWhenUnset(t *testing.T) {
	os.Unsetenv("TEST_KEY_UNSET")

	got := getEnv("TEST_KEY_UNSET", "fallback_value")
	want := "fallback_value"

	if got != want {
		t.Errorf("getEnv() = %q, want %q", got, want)
	}
}

func TestGetEnv_ReturnsFallbackWhenEmptyString(t *testing.T) {
	os.Setenv("TEST_KEY_EMPTY", "")
	defer os.Unsetenv("TEST_KEY_EMPTY")

	got := getEnv("TEST_KEY_EMPTY", "fallback_value")
	want := "fallback_value"

	if got != want {
		t.Errorf("getEnv() = %q, want %q", got, want)
	}
}
