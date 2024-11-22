// internal/db/migrations/loader.go
package migrations

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// LoadMigrations loads migrations from the given directory
func LoadMigrations(dir string) ([]Migration, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading migrations directory: %w", err)
	}

	migrations := make(map[int]Migration)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		if !strings.HasSuffix(filename, ".up.sql") {
			continue
		}

		version, name, err := parseMigrationFilename(filename)
		if err != nil {
			return nil, err
		}

		upContent, err := os.ReadFile(filepath.Join(dir, filename))
		if err != nil {
			return nil, fmt.Errorf("reading up migration: %w", err)
		}

		downFilename := fmt.Sprintf("%06d_%s.down.sql", version, name)
		downContent, err := os.ReadFile(filepath.Join(dir, downFilename))
		if err != nil {
			return nil, fmt.Errorf("reading down migration: %w", err)
		}

		migrations[version] = Migration{
			Version:    version,
			Name:       name,
			UpScript:   string(upContent),
			DownScript: string(downContent),
		}
	}

	if len(migrations) == 0 {
		return nil, fmt.Errorf("no migrations found in directory: %s", dir)
	}

	// Convert map to sorted slice
	result := make([]Migration, 0, len(migrations))
	for _, m := range migrations {
		result = append(result, m)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Version < result[j].Version
	})

	return result, nil
}

func parseMigrationFilename(filename string) (version int, name string, err error) {
	base := strings.TrimSuffix(filename, ".up.sql")
	parts := strings.SplitN(base, "_", 2)
	if len(parts) != 2 {
		return 0, "", fmt.Errorf("invalid migration filename format: %s", filename)
	}

	_, err = fmt.Sscanf(parts[0], "%d", &version)
	if err != nil {
		return 0, "", fmt.Errorf("invalid version number in filename: %s", filename)
	}

	name = parts[1]
	return version, name, nil
}
