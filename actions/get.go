package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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

	searchReq := anytype.SearchRequest{
		Query: "",
	}

	if len(tagList) > 0 && tagList[0] != "" {
		searchReq.Query = strings.Join(tagList, " ")
	}

	searchResp, err := client.Space(spaceID).Search(ctx, searchReq)
	if err != nil {
		log.Fatalf("Error searching objects: %v", err)
	}

	results := []map[string]string{}
	for _, obj := range searchResp.Data {
		content := obj.Name
		if obj.Snippet != "" {
			content = obj.Name + "\n" + obj.Snippet
		}

		var objTags []string
		for _, prop := range obj.Properties {
			if prop.Key == "tags" {
				for _, tag := range prop.MultiSelect {
					objTags = append(objTags, tag.Name)
				}
			}
		}

		codeBlock := extractCodeBlock(content)
		extractedTags := extractTags(content)
		allTags := append(objTags, extractedTags...)

		itemMap := map[string]string{
			"cmd":     codeBlock,
			"tags":    strings.Join(allTags, " "),
			"content": content,
			"name":    obj.ID,
		}
		results = append(results, itemMap)
	}

	jsonData, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		log.Fatalf("Error converting to JSON: %v", err)
	}

	fmt.Println(string(jsonData))
}
