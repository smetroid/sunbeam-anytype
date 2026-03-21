package transform

import (
	"encoding/json"
	"fmt"
	"os"
)

type Action struct {
	Type    string            `json:"type"`
	Title   string            `json:"title"`
	Text    string            `json:"text,omitempty"`
	Command string            `json:"command,omitempty"`
	Params  map[string]string `json:"params,omitempty"`
	Exit    *bool             `json:"exit,omitempty"`
}

type ListItem struct {
	Title       string   `json:"title"`
	Subtitle    string   `json:"subtitle,omitempty"`
	Accessories []string `json:"accessories,omitempty"`
	Actions     []Action `json:"actions"`
}

type DetailView struct {
	Markdown string   `json:"markdown"`
	Actions  []Action `json:"actions"`
}

type Manifest struct {
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Preferences []Preference `json:"preferences"`
	Commands    []Command    `json:"commands"`
}

type Preference struct {
	Name  string `json:"name"`
	Title string `json:"title"`
	Type  string `json:"type"`
}

type Command struct {
	Name  string `json:"name"`
	Title string `json:"title"`
	Mode  string `json:"mode"`
	Exit  string `json:"exit,omitempty"`
}

func GenerateManifest() Manifest {
	return Manifest{
		Title:       "Anytype",
		Description: "Search Anytype objects",
		Preferences: []Preference{
			{Name: "anytype_app_key", Title: "Anytype App Key", Type: "string"},
			{Name: "anytype_space_id", Title: "Anytype Space ID", Type: "string"},
		},
		Commands: []Command{
			{Name: "anytype-cmd", Title: "Anytype cmd blocks", Mode: "filter"},
			{Name: "anytype-snippet", Title: "Anytype snippets blocks", Mode: "filter"},
			{Name: "anytype-all", Title: "ALL objects", Mode: "filter"},
			{Name: "run-command", Title: "execute command", Mode: "tty", Exit: "true"},
			{Name: "view-command", Title: "view command", Mode: "detail", Exit: "false"},
			{Name: "edit-object", Title: "edit object", Mode: "tty", Exit: "false"},
		},
	}
}

func TransformItems(objects []map[string]string, templateType TemplateType) map[string]interface{} {
	var items []ListItem

	for _, obj := range objects {
		item := applyTemplate(obj, templateType)
		items = append(items, item)
	}

	result := map[string]interface{}{
		"items": items,
		"actions": []Action{
			{Type: "reload", Title: "Refresh items", Exit: boolPtr(true)},
		},
	}

	return result
}

type TemplateType int

const (
	CmdTemplate TemplateType = iota
	SnippetTemplate
	AllTemplate
)

func applyTemplate(obj map[string]string, templateType TemplateType) ListItem {
	var item ListItem

	switch templateType {
	case CmdTemplate:
		item = ListItem{
			Title:       obj["cmd"],
			Accessories: []string{obj["tags"]},
			Actions: []Action{
				{Type: "copy", Title: "Copy to clipboard", Text: obj["cmd"], Exit: boolPtr(true)},
				{Type: "run", Title: "Run Command", Command: "run-command", Params: map[string]string{"codeblock": obj["cmd"]}},
				{Type: "run", Title: "View Command", Command: "view-command", Params: map[string]string{"content": obj["content"], "codeblock": obj["cmd"]}},
			},
		}
	case SnippetTemplate:
		item = ListItem{
			Title:       obj["content"],
			Accessories: []string{obj["tags"]},
			Actions: []Action{
				{Type: "run", Title: "view cmd", Command: "view-command", Params: map[string]string{"content": obj["content"], "codeblock": obj["cmd"]}},
			},
		}
	case AllTemplate:
		subtitle := obj["cmd"]
		item = ListItem{
			Title:       obj["content"],
			Subtitle:    subtitle,
			Accessories: []string{obj["tags"]},
			Actions: []Action{
				{Type: "run", Title: "view object", Command: "view-command", Params: map[string]string{"content": obj["content"], "codeblock": obj["cmd"], "name": obj["name"]}},
				{Type: "run", Title: "edit object", Command: "edit-object", Params: map[string]string{"content": obj["content"], "codeblock": obj["cmd"], "name": obj["name"]}},
			},
		}
	}

	return item
}

func GenerateDetailView(params map[string]string) DetailView {
	return DetailView{
		Markdown: params["content"],
		Actions: []Action{
			{Type: "copy", Title: "Copy to clipboard", Text: params["codeblock"], Exit: boolPtr(false)},
			{Type: "run", Title: "Edit", Command: "edit-object", Params: map[string]string{"content": params["content"], "codeblock": params["codeblock"], "name": params["name"]}},
		},
	}
}

func (d DetailView) Map() map[string]interface{} {
	return map[string]interface{}{
		"markdown": d.Markdown,
		"actions":  d.Actions,
	}
}

func OutputJSON(data interface{}) {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
		return
	}
	fmt.Print(string(jsonBytes))
}

func boolPtr(b bool) *bool {
	return &b
}
