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

func (s *Storage) GetCategoryById(_ context.Context, id int64) (*models.Category, error) {
	const op = "storage.sqlite.GetCategoryById"
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

func (s *Storage) GetCategoryByName(_ context.Context, name string) (*models.Category, error) {
	const op = "storage.sqlite.GetCategoryByName"
	stmt, err := s.db.Prepare("SELECT id, name FROM Categories WHERE name =?")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()
	var category models.Category
	err = stmt.QueryRow(name).Scan(&category.ID, &category.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
	}
	return &category, nil
}

func (s *Storage) ListCategoriesReport(ctx context.Context, filter string) ([]models.CategoryReport, error) {
	const op = "storage.sqlite.ListCategoriesReport"
	sql := `
	SELECT sum(amount) AS cat_amount, Categories.name as cat_name
	FROM Expenses JOIN Categories ON Expenses.category_id = Categories.id
	GROUP BY Categories.name;`

	if filter == "month" {
		sql = `
		SELECT sum(amount) AS cat_amount, Categories.name as cat_name
		FROM Expenses JOIN Categories ON Expenses.category_id = Categories.id
		WHERE strftime('%m', date) = strftime('%m', datetime('now'))
		GROUP BY Categories.name;`
	}
	stmt, err := s.db.Prepare(sql)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var categories []models.CategoryReport

	rows, _ := stmt.QueryContext(ctx)

	for rows.Next() {
		var category models.CategoryReport
		err = rows.Scan(&category.Amount, &category.Name)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		categories = append(categories, category)
	}

	return categories, nil
}

func (s *Storage) Total() (int64, error) {
	const op = "storage.sqlite.TotalAmount"

	stmt, err := s.db.Prepare("SELECT sum(amount) FROM Expenses")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close() // nolint: errcheck

	var totalAmount int64

	err = stmt.QueryRow().Scan(&totalAmount)

	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return totalAmount, nil
}

func (s *Storage) SaveExpense(ctx context.Context, expense models.Expense) error {
	const op = "storage.sqlite.SaveExpense"

	category, _ := s.GetCategoryByName(ctx, expense.Category)

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
	stmt, err := s.db.Prepare("SELECT id, date(date) as date, description, amount, category_id FROM Expenses WHERE date(date) IS NOT NULL")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	defer stmt.Close() // nolint: errcheck

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
		cat, _ := s.GetCategoryById(ctx, expense.CategoryID)

		expense.Category = cat.Name
		expenses = append(expenses, expense)
	}

	return expenses, nil
}
func (s *Storage) ListCategories(ctx context.Context) ([]models.Category, error) {
	const op = "storage.sqlite.CategoriesList"
	stmt, err := s.db.Prepare("SELECT id, name FROM Categories")

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	defer stmt.Close()

	var categories []models.Category
	rows, err := stmt.QueryContext(ctx) // nolint: errcheck

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	defer rows.Close()

	for rows.Next() {
		var category models.Category
		err = rows.Scan(&category.ID, &category.Name)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		categories = append(categories, category)
	}
	return categories, nil
}
