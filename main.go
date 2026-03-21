package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
	"sunbeam-anytype/actions"
	"sunbeam-anytype/sunbeam"
)

type Payload struct {
	Command string            `json:"command"`
	Params  map[string]string `json:"params"`
	Query   string            `json:"query"`
}

type Config struct {
	Title       string         `yaml:"title"`
	Description string         `yaml:"description"`
	Preferences []Preference   `yaml:"preferences"`
	Commands    []Command      `yaml:"commands"`
	Tmpl        TemplateConfig `yaml:"templates"`
}

type Preference struct {
	Name  string `yaml:"name"`
	Title string `yaml:"title"`
	Type  string `yaml:"type"`
}

type Command struct {
	Name      string `yaml:"name"`
	Title     string `yaml:"title"`
	Mode      string `yaml:"mode"`
	Exit      string `yaml:"exit"`
	TagFilter string `yaml:"tag_filter"`
	Terminal  string `yaml:"terminal"`
	Editor    string `yaml:"editor"`
}

// Template definitions from YAML
type TemplateConfig struct {
	Items  map[string]ItemTemplate
	Detail DetailActionTemplate `yaml:"detail_template"`
	Global []Action             `yaml:"global_actions"`
}

type ItemTemplate struct {
	Title       string   `yaml:"title"`
	Subtitle    string   `yaml:"subtitle"`
	Accessories string   `yaml:"accessories"`
	Actions     []Action `yaml:"actions"`
}

type DetailActionTemplate struct {
	Markdown string   `yaml:"markdown"`
	Actions  []Action `yaml:"actions"`
}

var configCache *Config
var templateCache *TemplateConfig
var configModTime int64

func main() {
	// Auto-detect mode: JSON = sunbeam, flags = CLI
	if len(os.Args) < 2 {
		showManifest()
		return
	}

	if strings.HasPrefix(os.Args[1], "{") {
		handleSunbeamMode()
	} else {
		handleCLIMode()
	}
}

// ============ Sunbeam Mode ============

func handleSunbeamMode() {
	var payload Payload
	if err := json.Unmarshal([]byte(os.Args[1]), &payload); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing payload: %v\n", err)
		os.Exit(1)
	}

	switch payload.Command {
	case "anytype-cmd":
		handleAnytypeCmd()
	case "anytype-snippet":
		handleAnytypeSnippet()
	case "anytype-all":
		handleAnytypeAll()
	case "run-command":
		handleRunCommand(payload.Params)
	case "view-command":
		handleViewCommand(payload.Params)
	case "edit-object":
		handleEditObject(payload.Params)
	default:
		showManifest()
	}
}

func handleAnytypeCmd() {
	objects, err := getAnytypeObjects("cmd")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	result := TransformItems(objects, CmdTemplate)
	OutputJSON(result)
}

func handleAnytypeSnippet() {
	objects, err := getAnytypeObjects("snippet")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	result := TransformItems(objects, SnippetTemplate)
	OutputJSON(result)
}

func handleAnytypeAll() {
	objects, err := getAnytypeObjects("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	result := TransformItems(objects, AllTemplate)
	OutputJSON(result)
}

func handleRunCommand(params map[string]string) {
	codeblock := params["codeblock"]
	if codeblock == "" {
		fmt.Fprintf(os.Stderr, "No codeblock provided\n")
		os.Exit(1)
	}
	cmd := exec.Command("konsole", "-e", "bash", "-c", codeblock+"; exec bash")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running command: %v\n", err)
		os.Exit(1)
	}
}

func handleViewCommand(params map[string]string) {
	detail := GenerateDetailView(params)
	OutputJSON(detail.Map())
}

func handleEditObject(params map[string]string) {
	name := params["name"]
	content := params["content"]
	if name == "" {
		fmt.Fprintf(os.Stderr, "No name provided\n")
		os.Exit(1)
	}

	if !strings.HasPrefix(name, "/tmp/") {
		name = "/tmp/" + name + ".md"
	}

	if err := os.WriteFile(name, []byte(content), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
		os.Exit(1)
	}

	beforeEdit, err := getFileModTime(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting file time: %v\n", err)
		os.Exit(1)
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}
	editorCmd := exec.Command(editor, name)
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr
	if err := editorCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running editor: %v\n", err)
		os.Exit(1)
	}

	afterEdit, err := getFileModTime(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting file time: %v\n", err)
		os.Exit(1)
	}

	if beforeEdit == afterEdit {
		fmt.Println("No changes made to the file.")
		fmt.Print("Press enter to exit, and go back to the app")
		fmt.Scanln()
		os.Exit(0)
	}

	objectID := filepath.Base(name)
	objectID = strings.TrimSuffix(objectID, ".md")

	err = actions.UpdateAnytypeObject(objectID, getSpaceID(), getAppKey())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to update object\n%v\n", err)
		fmt.Print("Press enter to exit, and go back to the app")
		fmt.Scanln()
		os.Exit(1)
	}

	fmt.Printf("Success: updated object\n")
	fmt.Print("Press enter to exit, and go back to the app")
	fmt.Scanln()
}

// ============ CLI Mode ============

func handleCLIMode() {
	configPath := filepath.Join(os.Getenv("HOME"), ".config", "sunbeam", "sunbeam.json")
	preferences, err := sunbeam.ReadSunbeamConfig(configPath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	appKey := ""
	spaceID := ""
	if preferences.AnytypeAppKey == "" || preferences.SpaceID == "" {
		fmt.Printf("Warning: no values found in sunbeam configuration, trying environment variables\n")
		appKey = os.Getenv("ANYTYPE_APP_KEY")
		spaceID = os.Getenv("ANYTYPE_SPACE_ID")
	} else {
		appKey = preferences.AnytypeAppKey
		spaceID = preferences.SpaceID
	}

	if appKey == "" || spaceID == "" {
		fmt.Println("Error: ANYTYPE_APP_KEY and ANYTYPE_SPACE_ID must be set.")
		os.Exit(1)
	}

	tags := flag.String("tags", "", "Comma-separated list of tags")
	clipboard := flag.Bool("clipboard", false, "Create object from clipboard")
	shellCommand := flag.Bool("shellCommand", false, "Create object from last shell command")
	update := flag.Bool("update", false, "Update object")
	name := flag.String("name", "", "Object ID to update")
	flag.Parse()

	if *update {
		if *name == "" {
			fmt.Println("Please provide an object id to update")
			os.Exit(1)
		}
		err := actions.UpdateAnytypeObject(*name, spaceID, appKey)
		if err != nil {
			fmt.Printf("Error updating object: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Updated object id: %s\n", *name)
	} else if *clipboard || *shellCommand {
		opts := actions.PostOptions{
			Clipboard:    *clipboard,
			ShellCommand: *shellCommand,
			Tags:         *tags,
			SpaceID:      spaceID,
			AppKey:       appKey,
		}
		objID, err := actions.PostAnytypeObject(opts)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Object created successfully! ID: %s\n", objID)
	} else {
		objects, err := actions.GetAnytypeObjects(*tags, spaceID, appKey)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		data, _ := json.MarshalIndent(objects, "", "  ")
		fmt.Print(string(data))
	}
}

// ============ YAML Config ============

func loadConfig() *Config {
	yamlPath := "anytype.yaml"

	info, err := os.Stat(yamlPath)
	if err != nil {
		return nil
	}

	currentModTime := info.ModTime().Unix()

	if configCache != nil && configModTime == currentModTime {
		return configCache
	}

	data, err := os.ReadFile(yamlPath)
	if err != nil {
		return nil
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Invalid YAML in %s: %v\n", yamlPath, err)
		os.Exit(1)
	}

	configCache = &cfg
	configModTime = currentModTime

	return &cfg
}

func getTemplates() TemplateConfig {
	// Parse YAML directly to get templates
	data, err := os.ReadFile("anytype.yaml")
	if err == nil {
		var raw struct {
			Templates struct {
				Cmd     ItemTemplate `yaml:"cmd"`
				Snippet ItemTemplate `yaml:"snippet"`
				All     ItemTemplate `yaml:"all"`
			} `yaml:"templates"`
			DetailTemplate DetailActionTemplate `yaml:"detail_template"`
			GlobalActions  []Action             `yaml:"global_actions"`
		}
		if err := yaml.Unmarshal(data, &raw); err == nil {
			tpls := make(map[string]ItemTemplate)
			if raw.Templates.Cmd.Title != "" {
				tpls["cmd"] = raw.Templates.Cmd
			}
			if raw.Templates.Snippet.Title != "" {
				tpls["snippet"] = raw.Templates.Snippet
			}
			if raw.Templates.All.Title != "" {
				tpls["all"] = raw.Templates.All
			}
			if len(tpls) > 0 || raw.DetailTemplate.Markdown != "" || len(raw.GlobalActions) > 0 {
				return TemplateConfig{
					Items:  tpls,
					Detail: raw.DetailTemplate,
					Global: raw.GlobalActions,
				}
			}
		}
	}
	// Fallback to default templates
	return getDefaultTemplates()
}

func getDefaultTemplates() TemplateConfig {
	return TemplateConfig{
		Items: map[string]ItemTemplate{
			"cmd": {
				Title:       "{{.Cmd}}",
				Accessories: "{{.Tags}}",
				Actions: []Action{
					{Type: "copy", Title: "Copy to clipboard", Text: "{{.Cmd}}", Exit: boolPtr(true)},
					{Type: "run", Title: "Run Command", Command: "run-command", Params: map[string]string{"codeblock": "{{.Cmd}}"}},
					{Type: "run", Title: "View Command", Command: "view-command", Params: map[string]string{"content": "{{.Content}}", "codeblock": "{{.Cmd}}"}},
				},
			},
			"snippet": {
				Title:       "{{.Content}}",
				Accessories: "{{.Tags}}",
				Actions: []Action{
					{Type: "run", Title: "view cmd", Command: "view-command", Params: map[string]string{"content": "{{.Content}}", "codeblock": "{{.Cmd}}"}},
				},
			},
			"all": {
				Title:       "{{.Content}}",
				Subtitle:    "{{.Cmd}}",
				Accessories: "{{.Tags}}",
				Actions: []Action{
					{Type: "run", Title: "view object", Command: "view-command", Params: map[string]string{"content": "{{.Content}}", "codeblock": "{{.Cmd}}", "name": "{{.Name}}"}},
					{Type: "run", Title: "edit object", Command: "edit-object", Params: map[string]string{"content": "{{.Content}}", "codeblock": "{{.Cmd}}", "name": "{{.Name}}"}},
				},
			},
		},
		Detail: DetailActionTemplate{
			Markdown: "{{.content}}",
			Actions: []Action{
				{Type: "copy", Title: "Copy to clipboard", Text: "{{.codeblock}}", Exit: boolPtr(false)},
				{Type: "run", Title: "Edit", Command: "edit-object", Params: map[string]string{"content": "{{.content}}", "codeblock": "{{.codeblock}}", "name": "{{.name}}"}},
			},
		},
		Global: []Action{
			{Type: "reload", Title: "Refresh items", Exit: boolPtr(true)},
		},
	}
}

func showManifest() {
	cfg := loadConfig()
	if cfg != nil {
		manifest := Manifest{
			Title:       cfg.Title,
			Description: cfg.Description,
			Preferences: convertPreferences(cfg.Preferences),
			Commands:    convertCommands(cfg.Commands),
		}
		OutputJSON(manifest)
	} else {
		OutputJSON(GenerateManifest())
	}
}

type Manifest struct {
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Preferences []Pref       `json:"preferences"`
	Commands    []CommandDef `json:"commands"`
}

type Pref struct {
	Name  string `json:"name"`
	Title string `json:"title"`
	Type  string `json:"type"`
}

type CommandDef struct {
	Name  string `json:"name"`
	Title string `json:"title"`
	Mode  string `json:"mode"`
	Exit  string `json:"exit,omitempty"`
}

func convertPreferences(prefs []Preference) []Pref {
	result := make([]Pref, len(prefs))
	for i, p := range prefs {
		result[i] = Pref{
			Name:  p.Name,
			Title: p.Title,
			Type:  p.Type,
		}
	}
	return result
}

func convertCommands(cmds []Command) []CommandDef {
	result := make([]CommandDef, len(cmds))
	for i, c := range cmds {
		result[i] = CommandDef{
			Name:  c.Name,
			Title: c.Title,
			Mode:  c.Mode,
			Exit:  c.Exit,
		}
	}
	return result
}

func GenerateManifest() Manifest {
	return Manifest{
		Title:       "Anytype",
		Description: "Search Anytype objects",
		Preferences: []Pref{
			{Name: "anytype_app_key", Title: "Anytype App Key", Type: "string"},
			{Name: "anytype_space_id", Title: "Anytype Space ID", Type: "string"},
		},
		Commands: []CommandDef{
			{Name: "anytype-cmd", Title: "Anytype cmd blocks", Mode: "filter"},
			{Name: "anytype-snippet", Title: "Anytype snippets blocks", Mode: "filter"},
			{Name: "anytype-all", Title: "ALL objects", Mode: "filter"},
			{Name: "run-command", Title: "execute command", Mode: "tty", Exit: "true"},
			{Name: "view-command", Title: "view command", Mode: "detail", Exit: "false"},
			{Name: "edit-object", Title: "edit object", Mode: "tty", Exit: "false"},
		},
	}
}

// ============ Transform Logic (moved from extension package) ============

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

type TemplateType int

const (
	CmdTemplate TemplateType = iota
	SnippetTemplate
	AllTemplate
)

func TransformItems(objects []map[string]string, templateType TemplateType) map[string]interface{} {
	templates := getTemplates()

	var templateKey string
	switch templateType {
	case CmdTemplate:
		templateKey = "cmd"
	case SnippetTemplate:
		templateKey = "snippet"
	case AllTemplate:
		templateKey = "all"
	}

	tpl, ok := templates.Items[templateKey]
	if !ok {
		// Fallback to defaults
		tpl = getDefaultTemplates().Items[templateKey]
	}

	var items []ListItem
	for _, obj := range objects {
		item := applyTemplateFromConfig(obj, tpl)
		items = append(items, item)
	}

	result := map[string]interface{}{
		"items":   items,
		"actions": templates.Global,
	}

	return result
}

func applyTemplateFromConfig(obj map[string]string, tpl ItemTemplate) ListItem {
	item := ListItem{
		Title:       applyTemplateString(tpl.Title, obj),
		Subtitle:    applyTemplateString(tpl.Subtitle, obj),
		Accessories: []string{applyTemplateString(tpl.Accessories, obj)},
		Actions:     transformActions(tpl.Actions, obj),
	}
	return item
}

func applyTemplateString(tmpl string, data map[string]string) string {
	if tmpl == "" {
		return ""
	}
	t, err := template.New("").Parse(tmpl)
	if err != nil {
		return tmpl
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return tmpl
	}
	return buf.String()
}

func transformActions(actions []Action, data map[string]string) []Action {
	result := make([]Action, len(actions))
	for i, a := range actions {
		result[i] = Action{
			Type:    a.Type,
			Title:   applyTemplateString(a.Title, data),
			Text:    applyTemplateString(a.Text, data),
			Command: a.Command,
			Params:  transformParams(a.Params, data),
			Exit:    a.Exit,
		}
	}
	return result
}

func transformParams(params map[string]string, data map[string]string) map[string]string {
	result := make(map[string]string)
	for k, v := range params {
		result[k] = applyTemplateString(v, data)
	}
	return result
}

func GenerateDetailView(params map[string]string) DetailView {
	templates := getTemplates()

	detailTpl := templates.Detail
	if detailTpl.Markdown == "" {
		// Fallback to defaults
		detailTpl = getDefaultTemplates().Detail
	}

	return DetailView{
		Markdown: applyTemplateString(detailTpl.Markdown, params),
		Actions:  transformActions(detailTpl.Actions, params),
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

// ============ Helpers ============

func getAnytypeObjects(tags string) ([]map[string]string, error) {
	spaceID := getSpaceID()
	appKey := getAppKey()
	return actions.GetAnytypeObjects(tags, spaceID, appKey)
}

func getSpaceID() string {
	configPath := filepath.Join(os.Getenv("HOME"), ".config", "sunbeam", "sunbeam.json")
	preferences, _ := sunbeam.ReadSunbeamConfig(configPath)
	if preferences.SpaceID != "" {
		return preferences.SpaceID
	}
	if spaceID := os.Getenv("ANYTYPE_SPACE_ID"); spaceID != "" {
		return spaceID
	}
	fmt.Fprintln(os.Stderr, "Error: No space ID configured")
	os.Exit(1)
	return ""
}

func getAppKey() string {
	configPath := filepath.Join(os.Getenv("HOME"), ".config", "sunbeam", "sunbeam.json")
	preferences, _ := sunbeam.ReadSunbeamConfig(configPath)
	if preferences.AnytypeAppKey != "" {
		return preferences.AnytypeAppKey
	}
	if appKey := os.Getenv("ANYTYPE_APP_KEY"); appKey != "" {
		return appKey
	}
	fmt.Fprintln(os.Stderr, "Error: No app key configured")
	os.Exit(1)
	return ""
}

func getFileModTime(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.ModTime().Unix(), nil
}
