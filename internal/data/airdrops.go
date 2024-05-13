package data

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"gitlab.com/distributed_lab/kit/pgdb"
)

const (
	TxStatusPending   = "pending"
	TxStatusCompleted = "completed"
	TxStatusFailed    = "failed"
)

const airdropsTable = "airdrops"

type Airdrop struct {
	ID        string    `db:"id"`
	Nullifier string    `db:"nullifier"`
	Address   string    `db:"address"`
	TxHash    *string   `db:"tx_hash"`
	Amount    string    `db:"amount"`
	Status    string    `db:"status"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type AirdropsQ struct {
	db       *pgdb.DB
	selector squirrel.SelectBuilder
}

func NewAirdropsQ(db *pgdb.DB) *AirdropsQ {
	return &AirdropsQ{
		db:       db,
		selector: squirrel.Select("*").From(airdropsTable),
	}
}

func (q *AirdropsQ) New() *AirdropsQ {
	return NewAirdropsQ(q.db)
}

func (q *AirdropsQ) Insert(p Airdrop) (*Airdrop, error) {
	var res Airdrop
	stmt := squirrel.Insert(airdropsTable).SetMap(map[string]interface{}{
		"nullifier": p.Nullifier,
		"address":   p.Address,
		"tx_hash":   p.TxHash,
		"amount":    p.Amount,
		"status":    p.Status,
	}).Suffix("RETURNING *")

	if err := q.db.Get(&res, stmt); err != nil {
		return nil, fmt.Errorf("insert airdrop %+v: %w", p, err)
	}

	return &res, nil
}

func (q *AirdropsQ) Update(id string, values map[string]any) error {
	stmt := squirrel.Update(airdropsTable).SetMap(values).Where(squirrel.Eq{"id": id})

	if err := q.db.Exec(stmt); err != nil {
		return fmt.Errorf("update airdrop status [id=%s values=%v]: %w", id, values, err)
	}

	return nil
}

func (q *AirdropsQ) Delete(id string) error {
	stmt := squirrel.Delete(airdropsTable).Where(squirrel.Eq{"id": id})

	if err := q.db.Exec(stmt); err != nil {
		return fmt.Errorf("delete airdrop [id=%s]: %w", id, err)
	}

	return nil
}

func (q *AirdropsQ) Transaction(fn func() error) error {
	return q.db.Transaction(fn)
}

func (q *AirdropsQ) Select() ([]Airdrop, error) {
	var res []Airdrop

	if err := q.db.Select(&res, q.selector); err != nil {
		return nil, fmt.Errorf("select airdrops: %w", err)
	}

	return res, nil
}

func (q *AirdropsQ) Get() (*Airdrop, error) {
	var res Airdrop

	err := q.db.Get(&res, q.selector)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get airdrop: %w", err)
	}

	return &res, nil
}

func (q *AirdropsQ) Limit(limit uint64) *AirdropsQ {
	q.selector = q.selector.Limit(limit)
	return q
}

func (q *AirdropsQ) FilterByNullifier(nullifier string) *AirdropsQ {
	q.selector = q.selector.Where(squirrel.Eq{"nullifier": nullifier})
	return q
}

func (q *AirdropsQ) FilterByStatus(status string) *AirdropsQ {
	q.selector = q.selector.Where(squirrel.Eq{"status": status})
	return q
}
