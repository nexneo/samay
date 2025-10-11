package data

import (
	"reflect"
	"testing"
)

func TestUniqueSortedTags(t *testing.T) {
	input := []string{"Foo", "bar", "foo", "BAZ", "bar", "   baz  "}
	got := uniqueSortedTags(input)
	want := []string{"bar", "BAZ", "Foo"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("uniqueSortedTags(%v) = %v, want %v", input, got, want)
	}
}

func TestExtractTags(t *testing.T) {
	content := "Working on #Samay and fixing the #CLI. Also touched #samay again."
	got := extractTags(content)
	want := []string{"CLI", "Samay"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("extractTags(%q) = %v, want %v", content, got, want)
	}
}
