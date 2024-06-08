package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"

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

func (s *Storage) ListCategoriesReport(ctx context.Context, filter string, month int, year int) ([]models.CategoryReport, error) {
	const op = "storage.sqlite.ListCategoriesReport"
	strMonth := fmt.Sprintf("%02d", month)
	strYear := fmt.Sprintf("%04d", year)

	sql := `
	SELECT sum(amount) AS cat_amount, Categories.name as cat_name, Categories.color as color
	FROM Expenses JOIN Categories ON Expenses.category_id = Categories.id
	WHERE strftime('%m', date) = $1 AND strftime('%Y', date) = $2
	GROUP BY Categories.name;`

	stmt, err := s.db.Prepare(sql)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var categories []models.CategoryReport
	rows, err := stmt.QueryContext(ctx, strMonth, strYear)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	for rows.Next() {
		var category models.CategoryReport
		err = rows.Scan(&category.Amount, &category.Name, &category.Color)

		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		categories = append(categories, category)
	}

	return categories, nil
}

func (s *Storage) Total(filter string, month int, year int) (int64, error) {
	const op = "storage.sqlite.TotalAmount"

	strMonth := fmt.Sprintf("%02d", month)
	strYear := fmt.Sprintf("%04d", year)

	sql := `SELECT COALESCE(sum(amount), 0) FROM Expenses WHERE strftime('%m', date) = $1 AND strftime('%Y', date) = $2`

	stmt, err := s.db.Prepare(sql)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close() // nolint: errcheck

	var totalAmount int64

	err = stmt.QueryRow(strMonth, strYear).Scan(&totalAmount)

	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return totalAmount, nil
}

func (s *Storage) MedianAndMiddle(ctx context.Context, month int, year int) (int64, int64, error) {
	const op = "storage.sqlite.Median"

	var sum int
	var rowsCount int

	strMonth := fmt.Sprintf("%02d", month)
	strYear := fmt.Sprintf("%04d", year)

	sql := `SELECT sum(amount) AS day_amount, strftime('%d', date) as day
	FROM Expenses
	WHERE strftime('%m', date) = $1 AND strftime('%Y', date) = $2
	GROUP BY day;`

	stmt, _ := s.db.Prepare(sql)

	rows, err := stmt.QueryContext(ctx, strMonth, strYear)

	if err != nil {
		return -1, -1, fmt.Errorf("%s: %w", op, err)
	}

	amounts := make([]int, 0)
	for rows.Next() {
		var daily models.DailyStats

		err = rows.Scan(&daily.Total, &daily.Date)

		if err != nil {
			return -1, -1, fmt.Errorf("%s: %w", op, err)
		}

		sum += daily.Total
		rowsCount++

		amounts = append(amounts, daily.Total)
	}

	if rowsCount == 0 {
		return -1, -1, nil
	}

	middle := int64(sum / rowsCount)

	sort.Ints(amounts)

	return middle, int64(amounts[len(amounts)/2]), nil
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

func (s *Storage) UpdateExpense(ctx context.Context, expense models.Expense) error {
	const op = "storage.sqlite.SaveExpense"

	category, _ := s.GetCategoryByName(ctx, expense.Category)

	if category == nil {
		category = &models.Category{
			ID:       0,
			Name:     "TEST",
			ImageURL: "",
		}
	}

	stmt, err := s.db.Prepare(`
		UPDATE Expenses
		SET date=?,
		description=?,
		amount=?, category_id=?
		WHERE id=?;
	`)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	fmt.Println(expense)

	_, err = stmt.ExecContext(ctx, expense.Date, expense.Description, expense.Amount, category.ID, expense.ID)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
			return fmt.Errorf("%s: %w", op, storage.ErrUserExists)
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) ListExpenses(ctx context.Context, category string, month, year int64) ([]models.Expense, int, error) {
	const op = "storage.sqlite.ListExpenses"
	sql := `
	SELECT e.id as id, date(date) as date, description, amount, category_id, c.color, $1, $2
	FROM Expenses e JOIN Categories c on e.category_id = c.id
	WHERE date(date) IS NOT NULL AND strftime('%m', date) = strftime('%m', 'now') AND strftime('%Y', date) = strftime('%Y', 'now')
	ORDER BY date DESC
	`
	total := 0

	if category != "" {
		sql = fmt.Sprintf(`
		SELECT e.id as id, date(date) as date, description, amount, category_id, c.color, '1', '2'
		FROM Expenses e JOIN Categories c on e.category_id = c.id
		WHERE date(date) IS NOT NULL AND category_id = 
		(SELECT id FROM Categories WHERE name = '%s')
		`, category)

		sql += `AND strftime('%m', date) = $1 AND strftime('%Y', date) = $2
		ORDER BY date DESC`
	}
	stmt, err := s.db.Prepare(sql) // nolint: errcheck, gosec
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", op, err)
	}

	defer stmt.Close() // nolint: errcheck

	var expenses []models.Expense
	rows, err := stmt.Query(fmt.Sprintf("%02d", month), fmt.Sprintf("%04d", year)) // nolint: errcheck

	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", op, err)
	}

	defer rows.Close() // nolint: errcheck

	for rows.Next() {
		var expense models.Expense
		var dummy string
		err = rows.Scan(&expense.ID, &expense.Date, &expense.Description, &expense.Amount, &expense.CategoryID, &expense.Color, &dummy, &dummy)
		if err != nil {
			return nil, 0, fmt.Errorf("%s: %w", op, err)
		}
		cat, _ := s.GetCategoryById(ctx, expense.CategoryID)
		total += int(expense.Amount)
		expense.Category = cat.Name
		expenses = append(expenses, expense)
	}

	return expenses, total, nil
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
