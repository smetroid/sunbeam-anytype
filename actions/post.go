package actions

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/epheo/anytype-go"
	_ "github.com/epheo/anytype-go/client"
)

func getLastShellCommand() (string, error) {
	historyFile := os.Getenv("HISTFILE")
	if historyFile == "" {
		historyFile = os.ExpandEnv("$HOME/.bash_history")
	}

	data, err := os.ReadFile(historyFile)
	if err != nil {
		return "", fmt.Errorf("failed to read history file: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) != "" {
			return lines[i], nil
		}
	}

	return "", fmt.Errorf("no commands found in history file")
}

func UpdateAnytypeObject(objectID string, spaceID string, appKey string) error {
	ctx := context.Background()
	client := anytype.NewClient(
		anytype.WithBaseURL("http://localhost:31009"),
		anytype.WithAppKey(appKey),
	)

	file := fmt.Sprintf("/tmp/%s.md", objectID)
	content, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("reading file %s: %v", file, err)
	}

	updateReq := anytype.UpdateObjectRequest{
		Name: string(content),
	}

	_, err = client.Space(spaceID).Object(objectID).Update(ctx, updateReq)
	if err != nil {
		return fmt.Errorf("updating object: %v", err)
	}
	return nil
}

type PostOptions struct {
	Clipboard    bool
	ShellCommand bool
	Tags         string
	SpaceID      string
	AppKey       string
}

func PostAnytypeObject(opts PostOptions) (string, error) {
	ctx := context.Background()
	client := anytype.NewClient(
		anytype.WithBaseURL("http://localhost:31009"),
		anytype.WithAppKey(opts.AppKey),
	)

	props, err := client.Space(opts.SpaceID).Properties().List(ctx)
	if err != nil {
		return "", fmt.Errorf("listing properties: %v", err)
	}

	var tagsPropID string
	var tagsPropKey string
	for _, prop := range props {
		if prop.Name == "Tags" || prop.Name == "Tag" || prop.Key == "tags" || prop.Key == "tag" {
			tagsPropID = prop.ID
			tagsPropKey = prop.Key
			break
		}
	}

	var content string
	if opts.Clipboard {
		cmd := exec.Command("sunbeam", "paste")
		out, err := cmd.Output()
		if err != nil {
			return "", fmt.Errorf("reading clipboard: %v", err)
		}
		content = string(out)
	} else if opts.ShellCommand {
		lastCommand, err := getLastShellCommand()
		if err != nil {
			return "", fmt.Errorf("retrieving last shell command: %v", err)
		}
		content = lastCommand
	} else {
		return "", fmt.Errorf("please specify clipboard or shellCommand")
	}

	content = strings.TrimSpace(content)

	if content == "" {
		return "", fmt.Errorf("no content to post")
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter additional tags (comma-separated): ")
	additionalTags, _ := reader.ReadString('\n')
	additionalTags = strings.TrimSpace(additionalTags)

	var allTags []string
	if opts.Tags != "" {
		allTags = append(allTags, strings.Split(opts.Tags, ",")...)
	}
	if additionalTags != "" {
		allTags = append(allTags, strings.Split(additionalTags, ",")...)
	}

	var tagKeys []string
	if tagsPropKey != "" && len(allTags) > 0 {
		tagResp, err := client.Space(opts.SpaceID).Property(tagsPropID).Tags().List(ctx)
		if err != nil {
			fmt.Printf("Warning: Failed to list existing tags: %v\n", err)
		} else {
			for _, requestedTag := range allTags {
				requestedTag = strings.TrimSpace(requestedTag)
				if requestedTag == "" {
					continue
				}
				found := false
				for _, t := range tagResp {
					if strings.EqualFold(t.Name, requestedTag) {
						tagKeys = append(tagKeys, t.Key)
						found = true
						break
					}
				}
				if !found {
					tagCreateReq := anytype.CreateTagRequest{
						Name:  requestedTag,
						Color: "grey",
					}
					tagRespNew, err := client.Space(opts.SpaceID).Property(tagsPropID).Tags().Create(ctx, tagCreateReq)
					if err != nil {
						fmt.Printf("Warning: Failed to create tag '%s': %v\n", requestedTag, err)
					} else if tagRespNew != nil && tagRespNew.Tag.Key != "" {
						tagKeys = append(tagKeys, tagRespNew.Tag.Key)
					}
				}
			}
		}
	}

	firstWord := strings.Split(content, " ")[0]

	firstWord = strings.TrimSpace(firstWord)

	var hashtags []string
	if firstWord != "" {
		hashtags = append(hashtags, "#"+firstWord)
	}

	for _, tag := range allTags {
		tag = strings.TrimSpace(tag)
		if tag != "" {
			hashtags = append(hashtags, "#"+tag)
		}
	}

	var markdownContent string
	markdownContent = fmt.Sprintf("```bash\n%s\n```", content)

	createReq := anytype.CreateObjectRequest{
		TypeKey: "note",
		Name:    content,
		Body:    markdownContent,
	}

	obj, err := client.Space(opts.SpaceID).Objects().Create(ctx, createReq)
	if err != nil {
		return "", fmt.Errorf("creating object: %v", err)
	}

	if tagsPropKey != "" && len(tagKeys) > 0 {
		updateReq := anytype.UpdateObjectRequest{
			Properties: []anytype.PropertyLinkWithValue{
				{
					Key:         tagsPropKey,
					MultiSelect: tagKeys,
				},
			},
		}
		_, err = client.Space(opts.SpaceID).Object(obj.Object.ID).Update(ctx, updateReq)
		if err != nil {
			return "", fmt.Errorf("updating tags: %v", err)
		}
	}

	return obj.Object.ID, nil
}
