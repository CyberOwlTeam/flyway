package flyway_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/CyberOwlTeam/flyway"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
)

const (
	defaultPostgresDbUsername = "test-user"
	defaultPostgresDbPassword = "test-password"
)

func TestFlyway_parseInvalidRequest(t *testing.T) {
	tests := []struct {
		name string
		opts []testcontainers.ContainerCustomizer
	}{
		{
			name: "missing database url",
			opts: []testcontainers.ContainerCustomizer{
				testcontainers.WithImage(flyway.BuildFlywayImageVersion()),
				flyway.WithUser(defaultPostgresDbUsername),
				flyway.WithPassword(defaultPostgresDbPassword),
				flyway.WithMigrations(filepath.Join("testdata", flyway.DefaultMigrationsPath)),
			},
		},
		{
			name: "missing user",
			opts: []testcontainers.ContainerCustomizer{
				testcontainers.WithImage(flyway.BuildFlywayImageVersion()),
				flyway.WithDatabaseUrl("jdbc:postgresql://localhost:5432/test_db?sslmode=disable"),
				flyway.WithPassword(defaultPostgresDbPassword),
				flyway.WithMigrations(filepath.Join("testdata", flyway.DefaultMigrationsPath)),
			},
		},
		{
			name: "missing password",
			opts: []testcontainers.ContainerCustomizer{
				testcontainers.WithImage(flyway.BuildFlywayImageVersion()),
				flyway.WithDatabaseUrl("jdbc:postgresql://localhost:5432/test_db?sslmode=disable"),
				flyway.WithUser(defaultPostgresDbUsername),
				flyway.WithMigrations(filepath.Join("testdata", flyway.DefaultMigrationsPath)),
			},
		},
		{
			name: "missing migrations",
			opts: []testcontainers.ContainerCustomizer{
				testcontainers.WithImage(flyway.BuildFlywayImageVersion()),
				flyway.WithDatabaseUrl("jdbc:postgresql://localhost:5432/test_db?sslmode=disable"),
				flyway.WithUser(defaultPostgresDbUsername),
				flyway.WithPassword(defaultPostgresDbPassword),
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(tt *testing.T) {
			testCase := testCase

			flywayContainer, err := flyway.RunContainer(context.Background(),
				testCase.opts...,
			)

			require.Error(tt, err, "expected error")
			require.Nil(tt, flywayContainer, "expected nil container")
		})
	}
}
