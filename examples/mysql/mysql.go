package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/CyberOwlTeam/flyway"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	mysqlDBVersion  = "8.0.36"
	mysqlPort       = "3306"
	mysqlSrvName    = "mysql"
	mysqlDBName     = "mysqldb"
	mysqlDBUsername = "mysql-user"
	mysqlDBPassword = "password"
)

// mysqlContainer represents an abstration of MySQLContainer.
type mysqlContainer struct {
	*mysql.MySQLContainer
}

func main() {
	ctx := context.Background()

	// Create a new docker network
	nw, err := network.New(context.Background())
	if err != nil {
		log.Fatalln("failed to initiate network: ", err)
	}

	// Create a new MySQLContainer
	dbContainer, err := createTestMySQLContainer(ctx, nw)
	if err != nil {
		log.Fatalln("failed to create DB container: ", err)
	}
	defer func() {
		err = dbContainer.Terminate(ctx)
		if err != nil {
			log.Println("failed to terminate DB container: ", err)
		}
	}()

	// Create a Flyway container and run SQL migration
	flywayContainer, err := flyway.RunContainer(ctx,
		testcontainers.WithImage(flyway.BuildFlywayImageVersion()),
		network.WithNetwork([]string{"flyway"}, nw),
		flyway.WithDatabaseUrl(dbContainer.getNetworkURL()),
		flyway.WithUser(mysqlDBUsername),
		flyway.WithPassword(mysqlDBPassword),
		flyway.WithMigrations(filepath.Join("testdata", flyway.DefaultMigrationsPath)),
	)
	if err != nil {
		log.Fatalln("failed to run Flyway container: ", err)
	}
	defer func() {
		err = flywayContainer.Terminate(ctx)
		if err != nil {
			log.Println("failed to terminate Flyway container: ", err)
		}
	}()

	// Execute some queries on database
	err = execSampleQuery(ctx, dbContainer)
	if err != nil {
		log.Fatalln("failed to execute query", err)
	}

	// Inspect state of Flyway container
	state, err := flywayContainer.State(ctx)
	if err != nil {
		log.Fatalf("Flyway container exited with state [%#v] and error : %v", state, err)
	}
}

// execSampleQuery executes queries for dbContainer.
func execSampleQuery(ctx context.Context, dbContainer *mysqlContainer) error {
	uri, err := dbContainer.getExternalURL(ctx)
	if err != nil {
		return fmt.Errorf("get external URL: %w", err)
	}

	db, err := sql.Open("mysql", uri)
	if err != nil {
		return fmt.Errorf("open mysql conn: %w", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return fmt.Errorf("failed to ping DB: %w", err)
	}

	err = executeAsTransaction(db, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, "INSERT INTO stuff (name) VALUES ('test')")
		return err
	})

	if err != nil {
		return fmt.Errorf("failed to execute transaction: %w", err)
	}

	rows, err := db.Query("SELECT id, name, created_timestamp FROM stuff")
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id uuid.UUID
		var name string
		var created time.Time
		err := rows.Scan(&id, &name, &created)
		if err != nil {
			return fmt.Errorf("failed to scan: %w", err)
		}
	}

	err = rows.Err()
	if err != nil {
		return fmt.Errorf("rows: %w", err)
	}
	return nil
}

// executeAsTransaction wraps fUpdate so that panic can be recovered.
func executeAsTransaction(db *sql.DB, fUpdate func(*sql.Tx) error) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			err = fmt.Errorf("panic occurred in transaction: %v", p)
		} else if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()
	err = fUpdate(tx)
	return err
}

// getExternalURL returns the external URL to [mysqlContainer].
func (c *mysqlContainer) getExternalURL(ctx context.Context) (string, error) {
	url, err := c.ConnectionString(ctx)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s?parseTime=true", url), nil
}

// getNetworkURL returns the network URL to connect to [mysqlContainer].
func (c *mysqlContainer) getNetworkURL() string {
	return fmt.Sprintf("jdbc:mysql://%s:%s/%s?allowPublicKeyRetrieval=true", mysqlSrvName, mysqlPort, mysqlDBName)
}

// createTestMySQLContainer instantiates and runs a MySQL container.
func createTestMySQLContainer(ctx context.Context, nw *testcontainers.DockerNetwork) (*mysqlContainer, error) {
	port := fmt.Sprintf("%s/tcp", mysqlPort)
	dbContainer, err := mysql.RunContainer(ctx,
		network.WithNetwork([]string{mysqlSrvName}, nw),
		testcontainers.WithImage(fmt.Sprintf("mysql:%s", mysqlDBVersion)),
		mysql.WithDatabase(mysqlDBName),
		mysql.WithUsername(mysqlDBUsername),
		mysql.WithPassword(mysqlDBPassword),
		testcontainers.WithConfigModifier(func(config *container.Config) {
			config.ExposedPorts = map[nat.Port]struct{}{
				nat.Port(port): {},
			}
		}),
		testcontainers.WithWaitStrategy(
			wait.ForLog("ready for connections").
				WithOccurrence(1).
				WithStartupTimeout(10*time.Second)),
	)
	if err != nil {
		return nil, err
	}

	return &mysqlContainer{
		dbContainer,
	}, nil
}
