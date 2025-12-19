package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

var eslintConfigFiles = []string{
	".eslintrc.json",
	".eslintrc.js",
	".eslintrc.cjs",
	".eslintrc.yaml",
	".eslintrc.yml",
	".eslintrc",
	"eslint.config.js",
	"eslint.config.mjs",
	"eslint.config.cjs",
}

var prettierConfigFiles = []string{
	".prettierrc",
	".prettierrc.json",
	".prettierrc.yml",
	".prettierrc.yaml",
	".prettierrc.json5",
	".prettierrc.js",
	".prettierrc.cjs",
	".prettierrc.mjs",
	".prettierrc.toml",
	"prettier.config.js",
	"prettier.config.cjs",
	"prettier.config.mjs",
}

const minimalBiomeConfig = `{
  "linter": {
    "enabled": true,
    "rules": {
      "recommended": true
    }
  }
}
`

type configLocation struct {
	dir         string
	hasEslint   bool
	hasPrettier bool
}

func main() {
	inputDir := flag.String("input", "", "Input directory to scan for ESLint/Prettier configs")
	dryRun := flag.Bool("dry-run", false, "Only show what would be done without actually doing it")
	flag.Parse()

	if *inputDir == "" {
		fmt.Println("Usage: biome_configurator -input <directory> [-dry-run]")
		os.Exit(1)
	}

	absInputDir, err := filepath.Abs(*inputDir)
	if err != nil {
		fmt.Printf("Error resolving input directory: %v\n", err)
		os.Exit(1)
	}

	locations, err := findConfigs(absInputDir)
	if err != nil {
		fmt.Printf("Error scanning directory: %v\n", err)
		os.Exit(1)
	}

	if len(locations) == 0 {
		fmt.Println("No ESLint or Prettier config files found")
		return
	}

	fmt.Printf("Found configs in %d location(s):\n", len(locations))
	for dir, loc := range locations {
		flags := []string{}
		if loc.hasEslint {
			flags = append(flags, "eslint")
		}
		if loc.hasPrettier {
			flags = append(flags, "prettier")
		}
		fmt.Printf("  - %s [%s]\n", dir, strings.Join(flags, ", "))
	}

	for dir, loc := range locations {
		if *dryRun {
			fmt.Printf("\n[DRY RUN] Would migrate in: %s\n", dir)
			if loc.hasEslint {
				fmt.Printf("[DRY RUN]   - ESLint migration\n")
			}
			if loc.hasPrettier {
				fmt.Printf("[DRY RUN]   - Prettier migration\n")
			}
			continue
		}

		fmt.Printf("\nMigrating: %s\n", dir)

		biomeConfigPath := filepath.Join(dir, "biome.json")
		existingBiome := false
		if _, err := os.Stat(biomeConfigPath); err == nil {
			existingBiome = true
		}

		if !existingBiome {
			if err := os.WriteFile(biomeConfigPath, []byte(minimalBiomeConfig), 0o644); err != nil {
				fmt.Printf("Error creating biome.json: %v\n", err)
				continue
			}
		}

		migrationFailed := false

		if loc.hasEslint {
			if err := migrateEslintConfig(dir); err != nil {
				fmt.Printf("Error migrating ESLint config: %v\n", err)
				migrationFailed = true
			} else {
				fmt.Printf("  ✓ ESLint migrated\n")
			}
		}

		if loc.hasPrettier {
			if err := migratePrettierConfig(dir); err != nil {
				fmt.Printf("Error migrating Prettier config: %v\n", err)
				migrationFailed = true
			} else {
				fmt.Printf("  ✓ Prettier migrated\n")
			}
		}

		if migrationFailed && !existingBiome && !loc.hasEslint && !loc.hasPrettier {
			os.Remove(biomeConfigPath)
			continue
		}

		if err := patchBiomeConfig(biomeConfigPath); err != nil {
			fmt.Printf("Error patching biome.json: %v\n", err)
		}

		fmt.Printf("Created: %s\n", biomeConfigPath)
	}

	fmt.Println("\nDone! Make sure 'biome.json' is in your global gitignore:")
	fmt.Println("  echo 'biome.json' >> ~/.gitignore_global")
	fmt.Println("  git config --global core.excludesfile ~/.gitignore_global")
}

func findConfigs(root string) (map[string]*configLocation, error) {
	locations := make(map[string]*configLocation)

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsPermission(err) {
				return filepath.SkipDir
			}
			return err
		}

		if info.IsDir() {
			name := info.Name()
			if name == "node_modules" || name == ".git" || name == "dist" || name == "build" || name == ".devops" {
				return filepath.SkipDir
			}
			return nil
		}

		fileName := info.Name()
		dir := filepath.Dir(path)

		if slices.Contains(eslintConfigFiles, fileName) {
			if locations[dir] == nil {
				locations[dir] = &configLocation{dir: dir}
			}
			locations[dir].hasEslint = true
		}

		if slices.Contains(prettierConfigFiles, fileName) {
			if locations[dir] == nil {
				locations[dir] = &configLocation{dir: dir}
			}
			locations[dir].hasPrettier = true
		}

		return nil
	})

	return locations, err
}

func migrateEslintConfig(dir string) error {
	cmd := exec.Command("npx", "@biomejs/biome", "migrate", "eslint", "--write")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func migratePrettierConfig(dir string) error {
	cmd := exec.Command("npx", "@biomejs/biome", "migrate", "prettier", "--write")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func patchBiomeConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var config map[string]any
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	if formatter, ok := config["formatter"].(map[string]any); ok {
		formatter["formatWithErrors"] = true
	} else {
		config["formatter"] = map[string]any{
			"formatWithErrors": true,
		}
	}

	if js, ok := config["javascript"].(map[string]any); ok {
		js["parser"] = map[string]any{
			"unsafeParameterDecoratorsEnabled": true,
		}
	} else {
		config["javascript"] = map[string]any{
			"parser": map[string]any{
				"unsafeParameterDecoratorsEnabled": true,
			},
		}
	}

	output, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, output, 0o644)
}

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Biome Configurator - Migrate ESLint/Prettier configs to Biome\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  biome_configurator -input <directory> [options]\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nSupported ESLint config files:\n")
		fmt.Fprintf(os.Stderr, "  %s\n", strings.Join(eslintConfigFiles, ", "))
		fmt.Fprintf(os.Stderr, "\nSupported Prettier config files:\n")
		fmt.Fprintf(os.Stderr, "  %s\n", strings.Join(prettierConfigFiles, ", "))
	}
}
