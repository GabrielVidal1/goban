package config

import (
	"bufio"
	"os"
	"strings"
)

type Config struct {
	KanbanDir string
	Host      string
	Port      string
	AuthToken string
}

// Load reads .env, then applies flag overrides, env vars, and defaults.
// Pass empty strings for flagDir/flagHost/flagPort when no CLI flags are available.
func Load(flagDir, flagHost, flagPort string) *Config {
	loadDotEnv(".env")

	dir := flagDir
	if dir == "" {
		dir = os.Getenv("KANBAN_DIR")
	}
	if dir == "" {
		dir = "./kanban"
	}

	host := flagHost
	if host == "" {
		host = os.Getenv("HOST")
	}
	if host == "" {
		host = "localhost"
	}

	port := flagPort
	if port == "" {
		port = os.Getenv("PORT")
	}
	if port == "" {
		port = "8080"
	}

	return &Config{
		KanbanDir: dir,
		Host:      host,
		Port:      port,
		AuthToken: os.Getenv("AUTH_TOKEN"),
	}
}

func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.Trim(strings.TrimSpace(val), `"'`)
		if os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}
}
