package format

import (
	"reflect"
	"testing"
)

func TestNormalizeToolInputSchema_DeduplicatesRequired(t *testing.T) {
	schema := map[string]any{
		"type":     "object",
		"required": []any{"action", "app", "element_index", "action"},
		"properties": map[string]any{
			"action": map[string]any{
				"type": "string",
			},
		},
	}

	got := NormalizeToolInputSchema(schema)
	want := map[string]any{
		"type":     "object",
		"required": []any{"action", "app", "element_index"},
		"properties": map[string]any{
			"action": map[string]any{
				"type": "string",
			},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("NormalizeToolInputSchema() = %#v, want %#v", got, want)
	}
}

func TestNormalizeToolInputSchema_DeduplicatesNestedRequired(t *testing.T) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"nested": map[string]any{
				"type":     "object",
				"required": []any{"a", "b", "a"},
			},
		},
	}

	got := NormalizeToolInputSchema(schema)
	want := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"nested": map[string]any{
				"type":     "object",
				"required": []any{"a", "b"},
			},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("NormalizeToolInputSchema() = %#v, want %#v", got, want)
	}
}

func TestNormalizeToolInputSchema_PreservesNonStringRequired(t *testing.T) {
	schema := map[string]any{
		"type":     "object",
		"required": []any{"action", 123, "action", true},
	}

	got := NormalizeToolInputSchema(schema)
	want := map[string]any{
		"type":     "object",
		"required": []any{"action", 123, true},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("NormalizeToolInputSchema() = %#v, want %#v", got, want)
	}
}

func TestNormalizeToolInputSchema_HandlesStringSliceRequired(t *testing.T) {
	schema := map[string]any{
		"type":     "object",
		"required": []string{"action", "app", "action"},
	}

	got := NormalizeToolInputSchema(schema)
	want := map[string]any{
		"type":     "object",
		"required": []string{"action", "app"},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("NormalizeToolInputSchema() = %#v, want %#v", got, want)
	}
}

func TestNormalizeToolInputSchema_NilSchema(t *testing.T) {
	if got := NormalizeToolInputSchema(nil); got != nil {
		t.Errorf("NormalizeToolInputSchema(nil) = %#v, want nil", got)
	}
}
