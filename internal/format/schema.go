package format

// NormalizeToolInputSchema normalizes a JSON schema used as a tool input_schema.
//
// It recursively walks the schema and ensures that the "required" array contains
// unique elements. Some upstream providers (e.g. DeepSeek) strictly validate JSON
// Schema and reject duplicate entries in "required", while OpenAI/Codex may
// generate schemas with repeated property names.
//
// The function preserves nil values and non-array "required" values as-is.
func NormalizeToolInputSchema(schema map[string]any) map[string]any {
	if schema == nil {
		return nil
	}

	result := make(map[string]any, len(schema))
	for k, v := range schema {
		switch k {
		case "required":
			result[k] = normalizeRequired(v)
		default:
			result[k] = normalizeSchemaValue(v)
		}
	}
	return result
}

// normalizeRequired deduplicates a JSON Schema "required" array.
func normalizeRequired(v any) any {
	switch arr := v.(type) {
	case []string:
		return uniqueStrings(arr)
	case []any:
		return uniqueAnyStrings(arr)
	default:
		return v
	}
}

// normalizeSchemaValue recursively normalizes nested schema values.
func normalizeSchemaValue(v any) any {
	switch val := v.(type) {
	case map[string]any:
		return NormalizeToolInputSchema(val)
	case []any:
		out := make([]any, len(val))
		for i, item := range val {
			out[i] = normalizeSchemaValue(item)
		}
		return out
	case []map[string]any:
		// Some schema builders produce strongly-typed arrays instead of []any.
		out := make([]map[string]any, len(val))
		for i, item := range val {
			out[i] = NormalizeToolInputSchema(item)
		}
		return out
	case []string:
		// String slices are not expected to contain nested schemas, but keep them as-is.
		out := make([]string, len(val))
		copy(out, val)
		return out
	default:
		return v
	}
}

// uniqueStrings returns a new slice with duplicate strings removed, preserving order.
func uniqueStrings(arr []string) []string {
	if len(arr) == 0 {
		return arr
	}
	seen := make(map[string]struct{}, len(arr))
	result := make([]string, 0, len(arr))
	for _, s := range arr {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		result = append(result, s)
	}
	return result
}

// uniqueAnyStrings deduplicates a []any array where elements are expected to be strings.
// Non-string elements are preserved as-is.
func uniqueAnyStrings(arr []any) []any {
	if len(arr) == 0 {
		return arr
	}
	seen := make(map[string]struct{}, len(arr))
	result := make([]any, 0, len(arr))
	for _, item := range arr {
		str, ok := item.(string)
		if !ok {
			result = append(result, item)
			continue
		}
		if _, exists := seen[str]; exists {
			continue
		}
		seen[str] = struct{}{}
		result = append(result, item)
	}
	return result
}
