package postgres

import (
	"context"
	"fmt"
	"sync"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresPool struct {
	pool *pgxpool.Pool
}

func (p *PostgresPool) Close() {
	p.pool.Close()
}

func (p *PostgresPool) SetSchema(schema string, ctx context.Context) error {
	command := "SET search_path TO " + schema + ";"
	_, err := p.pool.Exec(ctx, command)
	if err != nil {
		return fmt.Errorf("unable to access schema : %s, error: %w", schema, err)
	}
	return nil
}

func (p *PostgresPool) NewConnection(ctx context.Context) (*PostgresConn, error) {
	con, err := p.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	return &PostgresConn{con}, nil
}

func (p *PostgresPool) NewTransaction(ctx context.Context) (PostgresTx, error) {
	tx, err := p.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		txErr = err
	}
	txInstance = PostgresTx{tx}
	if txErr != nil {
		return txInstance, fmt.Errorf("unable to create new transaction: %w", txErr)
	}
	return txInstance, nil
}

type PostgresConn struct {
	con *pgxpool.Conn
}

func (c *PostgresConn) Release() {
	c.con.Release()
}

func (c *PostgresConn) SetSchema(schema string, ctx context.Context) error {
	command := "SET search_path TO " + schema + ";"
	_, err := c.con.Exec(ctx, command)
	if err != nil {
		return fmt.Errorf("unable to access schema : %s, error: %w", schema, err)
	}
	return nil
}

type PostgresTx struct {
	tx pgx.Tx
}

func (t *PostgresTx) Commit(ctx context.Context) error {
	err := t.tx.Commit(ctx)
	if err != nil {
		return err
	}
	return nil
}

func(t *PostgresTx) SetSchema(ctx context.Context, schema string) error {
	_, err := t.tx.Exec(ctx, fmt.Sprintf(`set search_path="%s"`, schema))
	if err != nil {
		return fmt.Errorf("unable to access schema : %s, error: %w", schema, err)
	}
	return nil
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
	pgInstance *PostgresPool
	pgOnce sync.Once
	conErr error
	txInstance PostgresTx
	txErr error
)

func NewConnectionPool(config *Config, ctx context.Context) (*PostgresPool, error) {
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
		pgInstance = &PostgresPool{db}
	})
	if conErr != nil {
		return pgInstance, fmt.Errorf("unable to create connection pool: %w", conErr)
	}
	return pgInstance, nil
}




