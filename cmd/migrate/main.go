package main

import (
	"fmt"
	"net/url"
	"os"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/pranavbh-9117/IMB/pkg/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = buildDatabaseURL(cfg.Database)
	}

	m, err := migrate.New("file://migrations", dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "migrate.New: %v\n", err)
		os.Exit(1)
	}
	defer m.Close()

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	isProd := cfg.App.Env == "production" || cfg.App.Env == "prod"

	switch cmd {
	case "up":
		n := -1
		if len(os.Args) >= 3 {
			if val, err := strconv.Atoi(os.Args[2]); err == nil {
				n = val
			}
		}
		if n == -1 {
			err = m.Up()
		} else {
			err = m.Steps(n)
		}
	case "down":
		if isProd && !hasConfirmFlag(os.Args) {
			fmt.Fprintln(os.Stderr, "ERROR: destructive command 'down' in production requires --confirm flag")
			os.Exit(1)
		}
		n := 1
		if len(os.Args) >= 3 && os.Args[2] != "--confirm" {
			if val, err := strconv.Atoi(os.Args[2]); err == nil {
				n = val
			}
		}
		err = m.Steps(-n)
	case "version":
		v, dirty, verErr := m.Version()
		if verErr != nil {
			fmt.Fprintf(os.Stderr, "version: %v\n", verErr)
			os.Exit(1)
		}
		fmt.Printf("version=%d dirty=%v\n", v, dirty)
	case "force":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "force requires a version number")
			os.Exit(1)
		}
		v, _ := strconv.Atoi(os.Args[2])
		err = m.Force(v)
	case "drop":
		if !hasConfirmFlag(os.Args) {
			fmt.Fprintln(os.Stderr, "ERROR: destructive command 'drop' requires --confirm flag")
			os.Exit(1)
		}
		err = m.Drop()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}

	if err != nil && err != migrate.ErrNoChange {
		fmt.Fprintf(os.Stderr, "migrate %s: %v\n", cmd, err)
		os.Exit(1)
	}
	if err == migrate.ErrNoChange {
		fmt.Println("No migrations to apply.")
	}
}

func buildDatabaseURL(dbCfg config.DatabaseConfig) string {
	userInfo := url.UserPassword(dbCfg.User, dbCfg.Password)
	hostPort := fmt.Sprintf("%s:%s", dbCfg.Host, dbCfg.Port)
	query := fmt.Sprintf("sslmode=%s", url.QueryEscape(dbCfg.SSLMode))
	return fmt.Sprintf("postgres://%s@%s/%s?%s", userInfo.String(), hostPort, url.PathEscape(dbCfg.Name), query)
}

func hasConfirmFlag(args []string) bool {
	for _, arg := range args {
		if arg == "--confirm" {
			return true
		}
	}
	return false
}

func printUsage() {
	fmt.Println("Usage: migrate [up [N] | down N [--confirm] | version | force V | drop --confirm]")
}
