package postgres

import (
	"context"
	"fmt"
	"sync"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgresdb struct {
	db *pgxpool.Pool
}

type PostgresTx struct {
	tx pgx.Tx
}

type Config struct {
	ImportPath string
	Host string
	Port string
	DBName string
	User string
	Password string
	Schema string
}



var (
	pgInstance *Postgresdb
	pgOnce sync.Once
	conErr error
	txInstance PostgresTx
	txErr error
)

func NewConnection(config *Config, ctx context.Context) (*Postgresdb, error) {
	pgOnce.Do(func() {
		configstr := "user=" + config.User +
					" password=" + config.Password +
					" host=" + config.Host +
					" port=" + config.Port +
					" dbname=" + config.DBName
		db, err := pgxpool.New(ctx, configstr)
		if err != nil {
			conErr = err
		}
		pgInstance = &Postgresdb{db}
	})
	if conErr != nil {
		return pgInstance, fmt.Errorf("unable to create connection pool: %w", conErr)
	}
	return pgInstance, nil
}

func CloseConnection(conn *Postgresdb) {
	conn.db.Close()
}

func NewTransaction(postgres *Postgresdb, ctx context.Context) (PostgresTx, error) {
	tx, err := postgres.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		txErr = err
	}
	// Rollback is safe to call even if the tx is already closed, so if
	// the tx commits successfully, this is a no-op
	//defer tx.Rollback(ctx)
	txInstance = PostgresTx{tx}
	if txErr != nil {
		return txInstance, fmt.Errorf("unable to create new transaction: %w", txErr)
	}
	return txInstance, nil
}

func SchemaSetonTx(ptx PostgresTx, ctx context.Context, schema string) error {
	_, err := ptx.tx.Exec(ctx, fmt.Sprintf(`set search_path="%s"`, schema))
	if err != nil {
		return fmt.Errorf("unable to access schema : %s, error: %w", schema, err)
	}
	return nil
}

func SchemaSetonCon(postgres Postgresdb, ctx context.Context, schema string) error {
	command := "SET search_path TO " + schema + ";"
	_, err := postgres.db.Exec(ctx, command)
	if err != nil {
		return fmt.Errorf("unable to access schema : %s, error: %w", schema, err)
	}
	return nil
}

func CommitTransaction(ptx PostgresTx, ctx context.Context) error {
	err := ptx.tx.Commit(ctx)
	if err != nil {
		return err
	}
	return nil
}




