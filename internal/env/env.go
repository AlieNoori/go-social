package env

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func LoadEnv(filename string) error {
	err := loadFile(filename, false)
	if err != nil {
		log.Print("can not load .env file")
		return err
	}
	return nil
}

func loadFile(filename string, overload bool) error {
	envMap, err := readFile(filename)
	if err != nil {
		return err
	}

	currentEnv := map[string]bool{}
	rawEnv := os.Environ()
	for _, rawEnvLine := range rawEnv {
		key := strings.Split(rawEnvLine, "=")[0]
		currentEnv[key] = true
	}

	for key, value := range envMap {
		if !currentEnv[key] || overload {
			_ = os.Setenv(key, value)
		}
	}

	return nil
}

func readFile(filename string) (map[string]string, error) {
	// cwd, _ := os.Getwd()
	// fp := filepath.Join(cwd, filename)
	// file, err := os.ReadFile(fp)
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %v", filename, err)
	}
	return parseFile(file)
}

func parseFile(file []byte) (map[string]string, error) {
	envs := make(map[string]string)
	rows := bytes.Split(bytes.TrimSpace(file), []byte("\n"))

	for _, row := range rows {
		if len(row) == 0 || row[0] == '#' { // skip comments and empty lines
			continue
		}

		if bytes.HasPrefix(row, []byte("export ")) {
			row = bytes.TrimPrefix(row, []byte("export "))
		}

		parts := bytes.SplitN(row, []byte("="), 2)
		if len(parts) != 2 {
			log.Printf("skipping malformed line: %s", row)
			continue
		}
		for i, part := range parts {
			if bytes.Contains(part, []byte("\"")) {
				part = bytes.ReplaceAll(part, []byte("\""), []byte(""))
				parts[i] = bytes.TrimSpace(part)
			}
		}

		name := string(parts[0])
		value := string(parts[1])
		envs[name] = value
	}
	return envs, nil
}

func GetString(key, fallback string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	return val
}

func GetInt(key string, fallback int) int {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	valAsInt, err := strconv.Atoi(val)
	if err != nil {
		log.Print(err)
		return fallback
	}
	return valAsInt
}

func GetDuration(key string, fallback time.Duration) time.Duration {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	duration, err := time.ParseDuration(val)
	if err != nil {
		return fallback
	}

	return duration
}

func GetBool(key string, fallback bool) bool {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	valAsBool, err := strconv.ParseBool(val)
	if err != nil {
		log.Print(err)
		return fallback
	}
	return valAsBool
}
