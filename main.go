package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/itchyny/github-migrator/github"
	"github.com/itchyny/github-migrator/migrator"
	"github.com/itchyny/github-migrator/repo"
)

const name = "github-migrator"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", name, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: %s <source> <target>", name)
	}
	mig, err := createMigrator(args[0], args[1])
	if err != nil {
		return err
	}
	return mig.Migrate()
}

func createGitHubClient(tokenEnv, endpointEnv string) (github.Client, error) {
	token := os.Getenv(tokenEnv)
	if token == "" {
		return nil, fmt.Errorf("GitHub token not found (specify %s)", tokenEnv)
	}
	endpoint := os.Getenv(endpointEnv)
	if endpoint == "" {
		endpoint = "https://api.github.com"
	}
	cli := github.New(token, endpoint)
	user, err := cli.GetUser()
	if err != nil {
		return nil, fmt.Errorf("%s (or you may want to set %s)", err, endpointEnv)
	}
	fmt.Printf("login succeeded: %s\n", user.Login)
	return cli, nil
}

func createMigrator(sourcePath, targetPath string) (migrator.Migrator, error) {
	sourceCli, err := createGitHubClient(
		"GITHUB_MIGRATOR_SOURCE_API_TOKEN",
		"GITHUB_MIGRATOR_SOURCE_API_ENDPOINT",
	)
	if err != nil {
		return nil, err
	}
	targetCli, err := createGitHubClient(
		"GITHUB_MIGRATOR_TARGET_API_TOKEN",
		"GITHUB_MIGRATOR_TARGET_API_ENDPOINT",
	)
	if err != nil {
		return nil, err
	}
	source := repo.New(sourceCli, sourcePath)
	target := repo.New(targetCli, targetPath)
	return migrator.New(source, target, createUserMapping()), nil
}

func createUserMapping() map[string]string {
	m := make(map[string]string)
	for _, src := range strings.Split(os.Getenv("GITHUB_MIGRATOR_USERS_MAPPING"), ",") {
		xs := strings.Split(strings.TrimSpace(src), ":")
		if len(xs) == 2 && len(xs[0]) > 0 && len(xs[1]) > 0 {
			m[xs[0]] = xs[1]
		}
	}
	return m
}
