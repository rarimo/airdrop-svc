package pg

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"gitlab.com/distributed_lab/kit/pgdb"
)

const participantsTable = "participants"

type Participant struct {
	Nullifier string    `db:"nullifier"`
	Address   string    `db:"address"`
	CreatedAt time.Time `db:"created_at"`
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

func (q *ParticipantsQ) Insert(nullifier, address string) error {
	stmt := squirrel.Insert(participantsTable).Columns("nullifier").Values(nullifier)

	if err := q.db.Exec(stmt); err != nil {
		return fmt.Errorf("insert participant %s: %w", nullifier, err)
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
