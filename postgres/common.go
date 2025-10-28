package postgres

import (
	"context"
	"fmt"
	"time"
	"sync"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresdb struct {
	db *pgxpool.Pool
}

type postgresTx struct {
	tx pgx.Tx
}

type Config struct {
	URL string
	ImportPath string
	ExportPath string
	DebugPath string
}

type Subject struct {
	ID int64
	RDFAbout string
	Type string
	ValidFrom time.Time
	ValidTo time.Time
	Predicates []Predicate
}

type Predicate struct {
	ID int64
	Type string
	Objects []Object
}

type Object struct {
	ID int64
	ValueStr string
	ValueFlt float64
	Resources []Resource
}

type Resource struct {
	ID int64
	SubjectID int64
}

var (
	pgInstance *postgresdb
	pgOnce sync.Once
	conErr error
	txInstance postgresTx
	txErr error
)

func newConnection(config Config, ctx context.Context) (*postgresdb, error) {
	pgOnce.Do(func() {
		db, err := pgxpool.New(ctx, config.URL)
		if err != nil {
			conErr = err
		}
		pgInstance = &postgresdb{db}
	})
	if conErr != nil {
		return pgInstance, fmt.Errorf("unable to create connection pool: %w", conErr)
	}
	return pgInstance, nil
}

func newTransaction(postgres *postgresdb, ctx context.Context) (postgresTx, error) {
	tx, err := postgres.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		txErr = err
	}
	// Rollback is safe to call even if the tx is already closed, so if
	// the tx commits successfully, this is a no-op
	//defer tx.Rollback(ctx)
	txInstance = postgresTx{tx}
	if txErr != nil {
		return txInstance, fmt.Errorf("unable to create new transaction: %w", txErr)
	}
	return txInstance, nil
}

func schemaSet(ptx postgresTx, ctx context.Context, schema string) error {
	_, err := ptx.tx.Exec(ctx, fmt.Sprintf(`set search_path="%s"`, schema))
	if err != nil {
		return fmt.Errorf("unable to access schema : %s, error: %w", schema, txErr)
	}
	return nil
}

func commitTransaction(ptx postgresTx, ctx context.Context) error {
	err := ptx.tx.Commit(ctx)
	if err != nil {
		return err
	}
	return nil
}

func InsertSubject(ptx postgresTx, subject *Subject, ctx context.Context) error {
	command := "CALL insert_subject(@subtype, @rdfid, @valid_from, @valid_to, @ret_id)"
	args := pgx.NamedArgs{
    		"subtype": subject.Type,
    		"rdfid": subject.RDFAbout,
			"valid_from": "2000-01-01 00:00:00+02",
			"valid_to": "2999-12-31 00:00:00+02",
			"ret_id": subject.ID,
	}
	_, err := ptx.tx.Exec(ctx, command, args)
	if err != nil {
		return err
	}
	return nil
}


