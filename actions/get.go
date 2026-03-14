package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/epheo/anytype-go"
	_ "github.com/epheo/anytype-go/client"
)

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

func GetAnytypeObjects(tags *string, spaceID string, appKey string) {
	ctx := context.Background()
	client := anytype.NewClient(
		anytype.WithBaseURL("http://localhost:31009"),
		anytype.WithAppKey(appKey),
	)

	tagList := strings.Split(*tags, ",")
	for i := range tagList {
		tagList[i] = strings.TrimSpace(tagList[i])
	}

	searchReq := anytype.SearchRequest{}

	applyFilter := false

	if len(tagList) > 0 && tagList[0] != "" {
		props, err := client.Space(spaceID).Properties().List(ctx)
		if err == nil {
			var tagsPropID string
			var tagsPropKey string
			for _, prop := range props {
				if prop.Name == "Tags" || prop.Name == "Tag" || prop.Key == "tags" || prop.Key == "tag" {
					tagsPropID = prop.ID
					tagsPropKey = prop.Key
					break
				}
			}

			if tagsPropKey != "" {
				tagResp, err := client.Space(spaceID).Property(tagsPropID).Tags().List(ctx)

				var tagKeys []string

				if err == nil {
					for _, requestedTag := range tagList {
						found := false
						for _, t := range tagResp {
							if strings.EqualFold(t.Name, requestedTag) {
								tagKeys = append(tagKeys, t.Key)
								found = true
								break
							}
						}
						if !found {
							tagKeys = append(tagKeys, requestedTag)
						}
					}
				} else {
					for _, requestedTag := range tagList {
						tagKeys = append(tagKeys, requestedTag)
					}
				}

				if len(tagKeys) > 0 {
					applyFilter = true
					filterExpr := &anytype.FilterExpression{
						Operator: anytype.FilterOperatorAnd,
						Conditions: []anytype.FilterItem{
							anytype.MultiSelectFilter(
								tagsPropKey,
								anytype.FilterConditionIn,
								tagKeys,
							),
						},
					}
					searchReq.Filters = filterExpr
				}
			}
		}
	}

	if len(tagList) > 0 && tagList[0] != "" && !applyFilter {
		fmt.Println("[]")
		return
	}

	searchResp, err := client.Space(spaceID).Search(ctx, searchReq)
	if err != nil {
		if strings.Contains(err.Error(), "bad_request") || strings.Contains(err.Error(), "failed to build expression") {
			fmt.Println("[]")
			return
		}
		fmt.Fprintf(os.Stderr, "Error searching objects: %v\n", err)
		os.Exit(1)
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

		itemMap := map[string]string{
			"cmd":     codeBlock,
			"tags":    strings.Join(objTags, " "),
			"content": strings.TrimSpace(content),
			"name":    obj.ID,
		}

		results = append(results, itemMap)
	}

	jsonData, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error converting to JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(jsonData))
}
