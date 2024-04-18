package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/mattn/go-sqlite3"

	"github.com/kochnevns/finances-backend/internal/models"
	"github.com/kochnevns/finances-backend/internal/storage"

	_ "github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Stop() error {
	return s.db.Close()
}

func (s *Storage) GetCategory(_ context.Context, id int64) (*models.Category, error) {
	const op = "storage.sqlite.GetCategory"
	stmt, err := s.db.Prepare("SELECT id, name FROM Categories WHERE id =?")

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	defer stmt.Close() // nolint: errcheck

	var category models.Category

	err = stmt.QueryRow(id).Scan(&category.ID, &category.Name)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
	}

	return &category, nil
}

func (s *Storage) SaveExpense(ctx context.Context, expense models.Expense) error {
	const op = "storage.sqlite.SaveExpense"

	category, err := s.GetCategory(ctx, expense.CategoryID)

	if category == nil {
		category = &models.Category{
			ID:       0,
			Name:     "TEST",
			ImageURL: "",
		}
	}

	stmt, err := s.db.Prepare("INSERT INTO Expenses(date, description, amount, category_id) VALUES(?,?,?,?)")

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.ExecContext(ctx, expense.Date, expense.Description, expense.Amount, category.ID)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
			return fmt.Errorf("%s: %w", op, storage.ErrUserExists)
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) ListExpenses(ctx context.Context) ([]models.Expense, error) {
	const op = "storage.sqlite.ListExpenses"
	stmt, err := s.db.Prepare("SELECT id, date, description, amount, category_id FROM Expenses")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			err = fmt.Errorf("%s: %w", op, err)
		}
	}(stmt) // nolint: errcheck

	var expenses []models.Expense
	rows, err := stmt.Query() // nolint: errcheck

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	defer rows.Close() // nolint: errcheck

	for rows.Next() {
		var expense models.Expense
		err = rows.Scan(&expense.ID, &expense.Date, &expense.Description, &expense.Amount, &expense.CategoryID)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		fmt.Println(expense)
		cat, _ := s.GetCategory(ctx, expense.CategoryID)

		expense.Category = cat.Name
		expenses = append(expenses, expense)
	}

	return expenses, nil
}
