package entities

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

var username_mapper map[string]string = nil

type Username string

func (u Username) Link() string {
	username := u.resolveUsername()

	return fmt.Sprintf("[@%s](tg://resolve?domain=%s)", username, username)
}

func (u Username) Tag() string {
	return fmt.Sprintf("@%s", u.resolveUsername())
}

func (u Username) resolveUsername() string {
	var err error
	if username_mapper == nil {
		username_mapper, err = parseYAML()
		if err != nil {
			log.Printf("Не удалось спарсить yaml: %s", err)

			return string(u)
		}
	}

	username, ok := username_mapper[string(u)]
	if !ok {
		return string(u)
	}

	return username
}

func parseYAML() (map[string]string, error) {
	filename := os.Getenv("GITHUB_USERNAME_MAPPER")
	if filename == "" {
		return nil, fmt.Errorf("Отсутствуют переменная окружения GITHUB_USERNAME_MAPPER")
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var result map[string]string
	if err := yaml.Unmarshal([]byte(data), &result); err != nil {
		return nil, err
	}

	return result, nil
}
