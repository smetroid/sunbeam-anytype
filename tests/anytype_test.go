package anytype_test

import (
	"encoding/json"
	"os"
	"reflect"
	"sort"
	"testing"

	"sunbeam-anytype/transform"
)

func loadFixture(t *testing.T, filename string) []map[string]string {
	t.Helper()

	data, err := os.ReadFile("fixtures/" + filename)
	if err != nil {
		t.Fatalf("Failed to read fixture %s: %v", filename, err)
	}

	var result []map[string]string
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to parse fixture %s: %v", filename, err)
	}
	return result
}

func loadExpected(t *testing.T, filename string) string {
	t.Helper()

	data, err := os.ReadFile("fixtures/" + filename)
	if err != nil {
		t.Fatalf("Failed to read expected file %s: %v", filename, err)
	}
	return string(data)
}

func jsonEqual(a, b string) bool {
	var aVal, bVal interface{}
	if err := json.Unmarshal([]byte(a), &aVal); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(b), &bVal); err != nil {
		return false
	}
	return reflect.DeepEqual(normalizeMap(aVal), normalizeMap(bVal))
}

func normalizeMap(v interface{}) interface{} {
	switch v := v.(type) {
	case map[string]interface{}:
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		result := make(map[string]interface{}, len(v))
		for _, k := range keys {
			result[k] = normalizeMap(v[k])
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = normalizeMap(item)
		}
		return result
	default:
		return v
	}
}

func TestManifest(t *testing.T) {
	manifest := transform.GenerateManifest()

	got, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal manifest: %v", err)
	}

	expected := loadExpected(t, "expected_manifest.json")

	if !jsonEqual(string(got), expected) {
		t.Errorf("Manifest mismatch:\nGot:\n%s\nExpected:\n%s", got, expected)
	}
}

func TestTransformAnytypeCmd(t *testing.T) {
	objects := loadFixture(t, "mock_cmd_tag.json")

	transformed := transform.TransformItems(objects, transform.CmdTemplate)

	got, err := json.MarshalIndent(transformed, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal transform: %v", err)
	}

	expected := loadExpected(t, "expected_anytype-cmd.json")

	if !jsonEqual(string(got), expected) {
		t.Errorf("Transform mismatch:\nGot:\n%s\nExpected:\n%s", got, expected)
	}
}

func TestTransformAnytypeSnippet(t *testing.T) {
	objects := loadFixture(t, "mock_snippet_tag.json")

	transformed := transform.TransformItems(objects, transform.SnippetTemplate)

	got, err := json.MarshalIndent(transformed, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal transform: %v", err)
	}

	expected := loadExpected(t, "expected_anytype-snippet.json")

	if !jsonEqual(string(got), expected) {
		t.Errorf("Transform mismatch:\nGot:\n%s\nExpected:\n%s", got, expected)
	}
}

func TestTransformAnytypeAll(t *testing.T) {
	objects := loadFixture(t, "mock_objects.json")

	transformed := transform.TransformItems(objects, transform.AllTemplate)

	got, err := json.MarshalIndent(transformed, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal transform: %v", err)
	}

	expected := loadExpected(t, "expected_anytype-all.json")

	if !jsonEqual(string(got), expected) {
		t.Errorf("Transform mismatch:\nGot:\n%s\nExpected:\n%s", got, expected)
	}
}

func TestViewCommand(t *testing.T) {
	params := map[string]string{
		"content":   "Check git repository status\n\n```bash\ngit status\n```",
		"codeblock": "git status",
		"name":      "obj-001",
	}

	detail := transform.GenerateDetailView(params)

	got, err := json.MarshalIndent(detail.Map(), "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal detail: %v", err)
	}

	expected := loadExpected(t, "expected_view-command.json")

	if !jsonEqual(string(got), expected) {
		t.Errorf("Detail view mismatch:\nGot:\n%s\nExpected:\n%s", got, expected)
	}
}

func TestItemActionsCopy(t *testing.T) {
	objects := loadFixture(t, "mock_cmd_tag.json")

	result := transform.TransformItems(objects, transform.CmdTemplate)

	resultJSON, _ := json.Marshal(result)
	var resultMap map[string]interface{}
	json.Unmarshal(resultJSON, &resultMap)

	items := resultMap["items"].([]interface{})
	if len(items) == 0 {
		t.Fatal("Expected at least one item")
	}

	firstItem := items[0].(map[string]interface{})
	actions := firstItem["actions"].([]interface{})
	if len(actions) == 0 {
		t.Fatal("Expected at least one action")
	}

	firstAction := actions[0].(map[string]interface{})
	if firstAction["type"] != "copy" {
		t.Errorf("Expected first action type to be 'copy', got %v", firstAction["type"])
	}
	if firstAction["title"] != "Copy to clipboard" {
		t.Errorf("Expected title 'Copy to clipboard', got %v", firstAction["title"])
	}
}

func TestRefreshAction(t *testing.T) {
	objects := loadFixture(t, "mock_objects.json")

	result := transform.TransformItems(objects, transform.CmdTemplate)

	resultJSON, _ := json.Marshal(result)
	var resultMap map[string]interface{}
	json.Unmarshal(resultJSON, &resultMap)

	actions := resultMap["actions"].([]interface{})
	if len(actions) == 0 {
		t.Fatal("Expected at least one action")
	}

	refreshAction := actions[0].(map[string]interface{})
	if refreshAction["title"] != "Refresh items" {
		t.Errorf("Expected 'Refresh items', got %v", refreshAction["title"])
	}
	if refreshAction["type"] != "reload" {
		t.Errorf("Expected type 'reload', got %v", refreshAction["type"])
	}
	if refreshAction["exit"] != true {
		t.Errorf("Expected exit true, got %v", refreshAction["exit"])
	}
}
