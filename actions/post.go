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

func UpdateAnytypeObject(objectID string, spaceID string, appKey string) {
	ctx := context.Background()
	client := anytype.NewClient(
		anytype.WithBaseURL("http://localhost:31009"),
		anytype.WithAppKey(appKey),
	)

	file := fmt.Sprintf("/tmp/%s.md", objectID)
	content, err := os.ReadFile(file)
	if err != nil {
		fmt.Printf("Error reading file %s: %v\n", file, err)
		os.Exit(1)
	}

	updateReq := anytype.UpdateObjectRequest{
		Name: string(content),
	}

	_, err = client.Space(spaceID).Object(objectID).Update(ctx, updateReq)
	if err != nil {
		fmt.Printf("Error updating object: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Updated object id: %s\n", objectID)
}

func PostAnytypeObject(clipboard *bool, shellCommand *bool, tags *string, spaceID string, appKey string) {
	ctx := context.Background()
	client := anytype.NewClient(
		anytype.WithBaseURL("http://localhost:31009"),
		anytype.WithAppKey(appKey),
	)

	var content string
	if *clipboard {
		cmd := exec.Command("sunbeam", "paste")
		out, err := cmd.Output()
		if err != nil {
			fmt.Printf("Error reading clipboard: %v\n", err)
			os.Exit(1)
		}
		content = string(out)
	} else if *shellCommand {
		lastCommand, err := getLastShellCommand()
		if err != nil {
			fmt.Printf("Error retrieving last shell command: %v\n", err)
			os.Exit(1)
		}
		content = lastCommand
	} else {
		fmt.Print("Unable to post object ... please specify --clipboard or --shellCommand")
		os.Exit(1)
	}

	content = strings.TrimSpace(content)

	if content == "" {
		fmt.Println("No content to post.")
		os.Exit(1)
	}

	fmt.Printf("content: %s \n", content)
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter additional tags (comma-separated): ")
	additionalTags, _ := reader.ReadString('\n')
	additionalTags = strings.TrimSpace(additionalTags)

	var allTags []string
	if *tags != "" {
		allTags = append(allTags, strings.Split(*tags, ",")...)
	}
	if additionalTags != "" {
		allTags = append(allTags, strings.Split(additionalTags, ",")...)
	}

	firstWord := strings.Split(content, " ")[0]

	var hashtags []string
	hashtags = append(hashtags, "#cmds")
	hashtags = append(hashtags, "#"+firstWord)

	for _, tag := range allTags {
		tag = strings.TrimSpace(tag)
		if tag != "" {
			hashtags = append(hashtags, "#"+tag)
		}
	}

	markdownContent := fmt.Sprintf("```bash\n%s\n```\n\n%s", content, strings.Join(hashtags, " "))

	createReq := anytype.CreateObjectRequest{
		TypeKey: "page",
		Name:    markdownContent,
		Body:    content,
	}

	obj, err := client.Space(spaceID).Objects().Create(ctx, createReq)
	if err != nil {
		fmt.Printf("Error creating object: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Object created successfully! ID: %s\n", obj.Object.ID)
}
