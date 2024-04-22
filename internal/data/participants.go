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
)

const participantsTable = "participants"

type Participant struct {
	Nullifier string    `db:"nullifier"`
	Address   string    `db:"address"`
	Status    string    `db:"status"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type ParticipantsQ struct {
	db       *pgdb.DB
	selector squirrel.SelectBuilder
}

func NewParticipantsQ(db *pgdb.DB) *ParticipantsQ {
	return &ParticipantsQ{
		db:       db,
		selector: squirrel.Select("*").From(participantsTable),
	}
}

func (q *ParticipantsQ) New() *ParticipantsQ {
	return NewParticipantsQ(q.db)
}

func (q *ParticipantsQ) Insert(p Participant) (*Participant, error) {
	var res Participant
	stmt := squirrel.Insert(participantsTable).SetMap(map[string]interface{}{
		"nullifier": p.Nullifier,
		"address":   p.Address,
		"status":    p.Status,
	}).Suffix("RETURNING *")

	if err := q.db.Get(&res, stmt); err != nil {
		return nil, fmt.Errorf("insert participant %+v: %w", p, err)
	}

	return &res, nil
}

func (q *ParticipantsQ) UpdateStatus(nullifier, status string) error {
	stmt := squirrel.Update(participantsTable).Set("status", status).Where(squirrel.Eq{"nullifier": nullifier})

	if err := q.db.Exec(stmt); err != nil {
		return fmt.Errorf("update participant status [nullifier=%s newStatus=%s]: %w", nullifier, status, err)
	}

	return nil
}

func (q *ParticipantsQ) Delete(nullifier string) error {
	stmt := squirrel.Delete(participantsTable).Where(squirrel.Eq{"nullifier": nullifier})

	if err := q.db.Exec(stmt); err != nil {
		return fmt.Errorf("delete participant [nullifier=%s]: %w", nullifier, err)
	}

	return nil
}

func (q *ParticipantsQ) Transaction(fn func() error) error {
	return q.db.Transaction(fn)
}

func (q *ParticipantsQ) Select() ([]Participant, error) {
	var res []Participant

	if err := q.db.Select(&res, q.selector); err != nil {
		return nil, fmt.Errorf("select participants: %w", err)
	}

	return res, nil
}

func (q *ParticipantsQ) Get(nullifier string) (*Participant, error) {
	var res Participant

	err := q.db.Get(&res, q.selector.Where(squirrel.Eq{"nullifier": nullifier}))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get participant: %w", err)
	}

	return &res, nil
}

func (q *ParticipantsQ) Limit(limit uint64) *ParticipantsQ {
	q.selector = q.selector.Limit(limit)
	return q
}

func (q *ParticipantsQ) FilterByStatus(status string) *ParticipantsQ {
	q.selector = q.selector.Where(squirrel.Eq{"status": status})
	return q
}
