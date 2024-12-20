package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type CastConfiguration struct {
	IsNew bool

	SettingsX int
	SettingsY int
	SettingsW int
	SettingsH int

	EQWindowX int
	EQWindowY int
	EQWindowW int
	EQWindowH int
}

// LoadCastConfig loads an shindow config file
func LoadCastConfig(path string) (*CastConfiguration, error) {
	fmt.Println("Loading config from", path)
	_, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &CastConfiguration{
				IsNew:     true,
				EQWindowX: 0,
				EQWindowY: 0,
				EQWindowW: 1920,
				EQWindowH: 1080,
			}, nil
		} else {
			return nil, fmt.Errorf("stat shindow.ini: %w", err)
		}

	}

	r, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open: %s", strings.TrimPrefix(err.Error(), "open shindow.ini: "))
	}
	defer r.Close()

	var config CastConfiguration

	reader := bufio.NewScanner(r)
	for reader.Scan() {
		line := reader.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		if strings.Contains(line, "=") {
			parts := strings.Split(line, "=")
			if len(parts) != 2 {
				continue
			}
			key := strings.ToLower(strings.TrimSpace(parts[0]))
			value := strings.TrimSpace(parts[1])
			switch key {
			case "settings_x":
				config.SettingsX, err = strconv.Atoi(value)
				if err != nil {
					return nil, fmt.Errorf("parse settings_x: %w", err)
				}
			case "settings_y":
				config.SettingsY, err = strconv.Atoi(value)
				if err != nil {
					return nil, fmt.Errorf("parse settings_y: %w", err)
				}
			case "settings_w":
				config.SettingsW, err = strconv.Atoi(value)
				if err != nil {

					return nil, fmt.Errorf("parse settings_w: %w", err)
				}
			case "settings_h":
				config.SettingsH, err = strconv.Atoi(value)
				if err != nil {
					return nil, fmt.Errorf("parse settings_h: %w", err)
				}
			case "eq_window_x":
				config.EQWindowX, err = strconv.Atoi(value)
				if err != nil {
					return nil, fmt.Errorf("parse eq_window_x: %w", err)
				}
			case "eq_window_y":
				config.EQWindowY, err = strconv.Atoi(value)
				if err != nil {
					return nil, fmt.Errorf("parse eq_window_y: %w", err)
				}
			case "eq_window_w":
				config.EQWindowW, err = strconv.Atoi(value)
				if err != nil {
					return nil, fmt.Errorf("parse eq_window_w: %w", err)
				}
			case "eq_window_h":
				config.EQWindowH, err = strconv.Atoi(value)
				if err != nil {
					return nil, fmt.Errorf("parse eq_window_h: %w", err)
				}

			default:
				return nil, fmt.Errorf("unknown key in shindow.ini: %s", key)
			}
		}
	}

	return &config, nil
}

// Save saves the config
func (c *CastConfiguration) Save() error {
	fi, err := os.Stat("shindow.ini")
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("stat shindow.ini: %w", err)
		}
		w, err := os.Create("shindow.ini")
		if err != nil {
			return fmt.Errorf("create shindow.ini: %w", err)
		}
		w.Close()
	}
	if fi != nil && fi.IsDir() {
		return fmt.Errorf("shindow.ini is a directory")
	}

	r, err := os.Open("shindow.ini")
	if err != nil {
		return fmt.Errorf("open: %s", strings.TrimPrefix(err.Error(), "open shindow.ini: "))
	}
	defer r.Close()

	tmpConfig := CastConfiguration{}

	out := ""
	reader := bufio.NewScanner(r)
	for reader.Scan() {
		line := reader.Text()
		if strings.HasPrefix(line, "#") {
			out += line + "\n"
			continue
		}
		if !strings.Contains(line, "=") {
			out += line + "\n"
			continue
		}
		parts := strings.Split(line, "=")
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		switch key {
		case "settings_x":
			if tmpConfig.SettingsX == 1 {
				continue
			}
			out += fmt.Sprintf("%s = %d\n", key, c.SettingsX)
			tmpConfig.SettingsX = 1
			continue
		case "settings_y":
			if tmpConfig.SettingsY == 1 {
				continue
			}

			out += fmt.Sprintf("%s = %d\n", key, c.SettingsY)
			tmpConfig.SettingsY = 1
			continue
		case "settings_w":
			if tmpConfig.SettingsW == 1 {
				continue
			}

			out += fmt.Sprintf("%s = %d\n", key, c.SettingsW)
			tmpConfig.SettingsW = 1
			continue
		case "settings_h":
			if tmpConfig.SettingsH == 1 {
				continue
			}

			out += fmt.Sprintf("%s = %d\n", key, c.SettingsH)
			tmpConfig.SettingsH = 1
			continue
		case "eq_window_x":
			if tmpConfig.EQWindowX == 1 {
				continue
			}

			out += fmt.Sprintf("%s = %d\n", key, c.EQWindowX)
			tmpConfig.EQWindowX = 1
			continue
		case "eq_window_y":
			if tmpConfig.EQWindowY == 1 {
				continue
			}

			out += fmt.Sprintf("%s = %d\n", key, c.EQWindowY)
			tmpConfig.EQWindowY = 1
			continue
		case "eq_window_w":
			if tmpConfig.EQWindowW == 1 {
				continue
			}

			out += fmt.Sprintf("%s = %d\n", key, c.EQWindowW)
			tmpConfig.EQWindowW = 1
			continue
		case "eq_window_h":
			if tmpConfig.EQWindowH == 1 {
				continue
			}

			out += fmt.Sprintf("%s = %d\n", key, c.EQWindowH)
			tmpConfig.EQWindowH = 1
			continue
		}

		line = fmt.Sprintf("%s = %s", key, value)
		out += line + "\n"
	}

	if tmpConfig.SettingsX != 1 {
		out += fmt.Sprintf("settings_x = %d\n", c.SettingsX)
	}

	if tmpConfig.SettingsY != 1 {
		out += fmt.Sprintf("settings_y = %d\n", c.SettingsY)
	}

	if tmpConfig.SettingsW != 1 {
		out += fmt.Sprintf("settings_w = %d\n", c.SettingsW)
	}

	if tmpConfig.SettingsH != 1 {

		out += fmt.Sprintf("settings_h = %d\n", c.SettingsH)
	}

	if tmpConfig.EQWindowX != 1 {
		out += fmt.Sprintf("eq_window_x = %d\n", c.EQWindowX)
	}

	if tmpConfig.EQWindowY != 1 {
		out += fmt.Sprintf("eq_window_y = %d\n", c.EQWindowY)
	}

	if tmpConfig.EQWindowW != 1 {
		out += fmt.Sprintf("eq_window_w = %d\n", c.EQWindowW)
	}

	if tmpConfig.EQWindowH != 1 {
		out += fmt.Sprintf("eq_window_h = %d\n", c.EQWindowH)
	}

	err = os.WriteFile("shindow.ini", []byte(out), 0644)
	if err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}
