package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sunbeam-anytype/actions"
	"sunbeam-anytype/sunbeam"
)

func main() {
	configPath := filepath.Join(os.Getenv("HOME"), ".config", "sunbeam", "sunbeam.json")
	preferences, err := sunbeam.ReadSunbeamConfig(configPath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	appKey := ""
	spaceID := ""
	if preferences.AnytypeAppKey == "" || preferences.SpaceID == "" {
		fmt.Printf("Error: no values found in sunbeam anytype extension configuration ... trying environment variables")

		appKey = os.Getenv("ANYTYPE_APP_KEY")
		spaceID = os.Getenv("ANYTYPE_SPACE_ID")
	} else {
		appKey = preferences.AnytypeAppKey
		spaceID = preferences.SpaceID
	}

	if appKey == "" || spaceID == "" {
		fmt.Println("Environment variables ANYTYPE_APP_KEY and ANYTYPE_SPACE_ID must be set.")
		os.Exit(1)
	}

	tags := flag.String("tags", "", "Comma-separated list of tags for the object (e.g., 'shell,commands')")
	clipboard := flag.Bool("clipboard", false, "Create an object using the contents of the clipboard")
	shellCommand := flag.Bool("shellCommand", false, "Create an object using the last shell command")
	update := flag.Bool("update", false, "Update object")
	name := flag.String("name", "", "id of object to update")
	flag.Parse()

	if *update {
		if *name == "" {
			fmt.Println("Please provide an object id to update")
			os.Exit(1)
		} else {
			actions.UpdateAnytypeObject(*name, spaceID, appKey)
		}
	} else if *clipboard || *shellCommand {
		actions.PostAnytypeObject(clipboard, shellCommand, tags, spaceID, appKey)
	} else {
		actions.GetAnytypeObjects(tags, spaceID, appKey)
	}
}
