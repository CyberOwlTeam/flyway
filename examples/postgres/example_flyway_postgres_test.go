package main

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/CyberOwlTeam/flyway"

	"github.com/testcontainers/testcontainers-go"
	tcnetwork "github.com/testcontainers/testcontainers-go/network"
)

func ExampleFlyway_postgres() {
	// runFlywayContainer {
	ctx := context.Background()
	nw, err := tcnetwork.New(ctx)
	if err != nil {
		log.Fatalf("failed to start network: %s", err) // nolint:gocritic
	}
	postgresContainer, err := createTestPostgresContainer(ctx, nw)
	if err != nil {
		log.Fatalf("failed to start postgres container: %s", err) // nolint:gocritic
	}

	flywayContainer, err := flyway.RunContainer(ctx,
		testcontainers.WithImage(flyway.BuildFlywayImageVersion()),
		tcnetwork.WithNetwork([]string{"flyway"}, nw),
		flyway.WithDatabaseUrl(postgresContainer.getNetworkUrl()),
		flyway.WithUser(defaultPostgresDbUsername),
		flyway.WithPassword(defaultPostgresDbPassword),
		flyway.WithConnectRetries(3),
		flyway.WithTable("my_schema_history"),
		flyway.WithGroup("my_group"),
		flyway.WithTimeout(1*time.Minute),
		flyway.WithMigrations(filepath.Join("testdata", flyway.DefaultMigrationsPath)),
	)
	if err != nil {
		log.Fatalf("failed to start container: %s", err) // nolint:gocritic
	}

	// Clean up the container
	defer func() {
		if err := flywayContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err) // nolint:gocritic
		}
	}()
	//}

	state, err := flywayContainer.State(ctx)
	if err != nil {
		log.Fatalf("failed to get container state: %s", err) // nolint:gocritic
	}

	fmt.Println(state.Running)  // the container should terminate immediately
	fmt.Println(state.ExitCode) // the exit code should be 0, as the flyway migrations were successful

	// Output:
	// false
	// 0
}
