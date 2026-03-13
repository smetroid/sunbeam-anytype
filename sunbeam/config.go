package sunbeam

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type AnytypeObject struct {
	ID      string   `json:"id"`
	Type    string   `json:"typeKey"`
	Name    string   `json:"name"`
	Content string   `json:"content"`
	Tags    []string `json:"tags"`
}

type Preferences struct {
	AnytypeAppKey string `json:"anytype_app_key"`
	SpaceID       string `json:"anytype_space_id"`
}

type AnytypeExtension struct {
	Origin      string      `json:"origin"`
	Preferences Preferences `json:"preferences"`
}

type Extensions struct {
	Anytype AnytypeExtension `json:"anytype"`
}

type SunbeamConfig struct {
	Extensions Extensions `json:"extensions"`
}

func ReadSunbeamConfig(configPath string) (Preferences, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return Preferences{}, fmt.Errorf("error opening configuration file: %v", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return Preferences{}, fmt.Errorf("error reading configuration file: %v", err)
	}

	var config SunbeamConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		return Preferences{}, fmt.Errorf("error parsing JSON: %v", err)
	}

	return config.Extensions.Anytype.Preferences, nil
}
