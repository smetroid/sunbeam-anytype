package actions

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/epheo/anytype-go"
	_ "github.com/epheo/anytype-go/client"
)

type AnytypeObject struct {
	Cmd     string `json:"cmd"`
	Tags    string `json:"tags"`
	Content string `json:"content"`
	Name    string `json:"name"`
}

func extractCodeBlock(content string) string {
	re := regexp.MustCompile("(?s)```\\w*\\n(.*?)\\n```")
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func extractTags(content string) []string {
	re := regexp.MustCompile(`(?i)(#[a-zA-Z0-9-_]+(?:\s*#[a-zA-Z0-9-_]+)*)`)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		tagsLine := matches[1]
		tags := strings.Fields(tagsLine)
		for i, tag := range tags {
			tags[i] = strings.TrimPrefix(tag, "#")
		}
		return tags
	}
	return nil
}

func GetAnytypeObjects(tags string, spaceID string, appKey string) ([]map[string]string, error) {
	ctx := context.Background()
	client := anytype.NewClient(
		anytype.WithBaseURL("http://localhost:31009"),
		anytype.WithAppKey(appKey),
	)

	tagList := strings.Split(tags, ",")
	for i := range tagList {
		tagList[i] = strings.TrimSpace(tagList[i])
	}

	searchReq := anytype.SearchRequest{}

	searchResp, err := client.Space(spaceID).Search(ctx, searchReq)
	if err != nil {
		return nil, fmt.Errorf("searching objects: %v", err)
	}

	results := []map[string]string{}
	for _, obj := range searchResp.Data {
		var markdown, content string

		objResp, err := client.Space(spaceID).Object(obj.ID).Get(ctx, anytype.WithFormat("md"))
		if err == nil && objResp.Object != nil {
			markdown = objResp.Object.Markdown
			content = objResp.Object.Markdown
			if content == "" {
				content = objResp.Object.Snippet
				if content == "" {
					content = objResp.Object.Name
				}
			}
		} else {
			markdown = obj.Markdown
			content = obj.Snippet
			if content == "" {
				content = obj.Name
			}
		}

		codeBlock := strings.TrimSpace(extractCodeBlock(markdown))

		var objTags []string
		for _, prop := range obj.Properties {
			if prop.Key == "tags" || prop.Key == "tag" {
				for _, tag := range prop.MultiSelect {
					objTags = append(objTags, tag.Name)
				}
				break
			}
		}

		if len(tagList) > 0 && tagList[0] != "" {
			hasTag := false
			for _, requestedTag := range tagList {
				for _, objTag := range objTags {
					if strings.EqualFold(objTag, requestedTag) {
						hasTag = true
						break
					}
				}
				if hasTag {
					break
				}
			}
			if !hasTag {
				continue
			}
		}

		itemMap := map[string]string{
			"cmd":     codeBlock,
			"tags":    strings.Join(objTags, " "),
			"content": strings.TrimSpace(content),
			"name":    obj.ID,
		}

		results = append(results, itemMap)
	}

	return results, nil
}
