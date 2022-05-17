package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gavinturner/VinylRetailChecker/util/cfg"
	"github.com/jmoiron/sqlx"
	"strings"
	"time"
)

func init() {}

type PostgresDBOpts struct {
	Host             string
	Port             string
	User             string
	Password         string
	Database         string
	Driver           string
	MaxOpenConns     int
	MaxIdleConns     int
	StatementTimeout time.Duration
}

const DefaultStatementTimeout = 5 * time.Minute

func NewPostgresDB(opts *PostgresDBOpts) (DB, error) {
	connStringVals := []string{
		fmt.Sprintf("host=%s", opts.Host),
		fmt.Sprintf("dbname=%s", opts.Database),
		fmt.Sprintf("user=%s", opts.User),
		"sslmode=disable",
		fmt.Sprintf("options='-c statement_timeout=%v'", int64(opts.StatementTimeout.Seconds())*1000),
	}

	if opts.Port != "" {
		connStringVals = append(connStringVals, fmt.Sprintf("port=%s", opts.Port))
	}

	if opts.Password != "" {
		connStringVals = append(connStringVals, fmt.Sprintf("password=%s", opts.Password))
	}

	db, err := sqlx.Open(opts.Driver, strings.Join(connStringVals, " "))
	if err != nil {
		return DB{}, fmt.Errorf("error connecting to psql database: %v", err)
	}

	if opts.MaxOpenConns > 0 {
		db.SetMaxOpenConns(opts.MaxOpenConns)
	}

	if opts.MaxIdleConns > 0 {
		db.SetMaxIdleConns(opts.MaxIdleConns)
	}
	return DB{db}, nil
}

func MustGetOptsWithMaxConns(maxOpenConns int, maxIdleConns int) *PostgresDBOpts {
	dbHost, _ := cfg.StringSetting("DB_HOST")
	dbPort, _ := cfg.StringSetting("DB_PORT")
	dbUser, _ := cfg.StringSetting("DB_USER")
	dbPassword, _ := cfg.StringSetting("DB_PASSWORD")
	dbDatabase, _ := cfg.StringSetting("DB_DATABASE")
	dbDriver, _ := cfg.StringSetting("DB_DRIVER")

	if dbHost == "" {
		panic("DB_HOST must be set")
	}
	if dbUser == "" {
		panic("DB_USER must be set")
	}
	if dbDatabase == "" {
		panic("DB_DATABASE must be set")
	}
	if dbDriver == "" {
		panic("DB_DRIVER must be set")
	}
	if dbPassword == "" {
		panic("DB_PASSWORD must be set")
	}

	return &PostgresDBOpts{
		Host:             dbHost,
		Port:             dbPort,
		User:             dbUser,
		Password:         dbPassword,
		Database:         dbDatabase,
		Driver:           dbDriver,
		MaxOpenConns:     maxOpenConns,
		MaxIdleConns:     maxIdleConns,
		StatementTimeout: DefaultStatementTimeout,
	}
}

// MustGetOpts gets database configuration from the environment.
// Panics if any configuration items are missing.
func MustGetOpts() *PostgresDBOpts {
	maxOpenConns, _ := cfg.IntSetting("DB_MAXCONNS")
	maxIdleConns, _ := cfg.IntSetting("DB_MAXIDLECONNS")
	return MustGetOptsWithMaxConns(maxOpenConns, maxIdleConns)
}

func LocalOpts() *PostgresDBOpts {
	dbHost, _ := cfg.StringSetting("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbPort, _ := cfg.StringSetting("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}
	dbUser, _ := cfg.StringSetting("DB_USER")
	if dbUser == "" {
		dbUser = "vinylretailers"
	}
	dbPassword, _ := cfg.StringSetting("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = ""
	}
	dbDatabase, _ := cfg.StringSetting("DB_DATABASE")
	if dbDatabase == "" {
		dbDatabase = "vinylretailers"
	}

	return &PostgresDBOpts{
		Host:     dbHost,
		Port:     dbPort,
		User:     dbUser,
		Password: dbPassword,
		Database: dbDatabase,
		Driver:   "postgres",
	}
}

// DB exposes all the methods available on the *sqlx.DB struct
type DB struct {
	*sqlx.DB
}

// Querier is an interface which is implemented by both *sqlx.DB and *sqlx.Tx
type Querier interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Select(dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	Get(dest interface{}, query string, args ...interface{}) error
	Rebind(query string) string
	QueryRow(query string, args ...interface{}) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

// Many packages use the same schema, this function
// knows which order to clear out tables so each package doesn't need to know
// (given foreign key constraints, etc)
func (db *DB) ResetDB() {
	db.ResetTables()
}

// Many packages use the same schema, this function
// knows which order to clear out tables so each package doesn't need to know
// (given foreign key constraints, etc)
func (db *DB) ResetTables() {

	preClearScripts(db)

	tablesInOrder := []string{
		"users_following",
		"artists",
		"users",
	}

	for _, tab := range tablesInOrder {
		db.MustExec(fmt.Sprintf("DELETE FROM %v", tab))
	}

	tablesWithAnIDThatNeedsResetting := []string{}

	for _, tab := range tablesWithAnIDThatNeedsResetting {
		db.MustExec(fmt.Sprintf("ALTER SEQUENCE %v_id_seq RESTART WITH 1;", tab))
	}
}

func preClearScripts(db *DB) {
	// Clear any custom questionnaire sections + types (need to do this before clearing orgs)
	/*
		db.MustExec(`
			DELETE FROM survey_type_sections WHERE survey_type_id IN
				(SELECT id FROM survey_types WHERE organisation_id IS NOT NULL)
		`)
	*/
}
