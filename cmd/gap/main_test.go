package main

import "testing"

func TestParseProjectSpecWithPositionalName(t *testing.T) {
	spec, err := parseProjectSpec([]string{"gap-test", "-module", "example.com/gap-test", "-force"})
	if err != nil {
		t.Fatal(err)
	}
	if spec.Name != "gap-test" {
		t.Fatalf("expected gap-test, got %q", spec.Name)
	}
	if spec.Module != "example.com/gap-test" {
		t.Fatalf("unexpected module: %q", spec.Module)
	}
	if !spec.Force {
		t.Fatal("expected force to be true")
	}
}

func TestParseProjectSpecRejectsTwoNameForms(t *testing.T) {
	_, err := parseProjectSpec([]string{"gap-test", "-name", "another-name"})
	if err == nil {
		t.Fatal("expected conflicting project names to fail")
	}
}
