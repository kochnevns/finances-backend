package finances

import (
	"context"
	"log/slog"

	financesgrpc "github.com/kochnevns/finances-backend/internal/grpc/finances"
	"github.com/kochnevns/finances-backend/internal/models"
)

type Finances struct {
	log                      *slog.Logger
	expenseSaver             ExpensesSaver
	expensesProvider         ExpensesProvider
	categoriesReportProvider CategoriesReportProvider
	categoriesProvider       CategoriesProvider
}

//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=URLSaver
type ExpensesSaver interface {
	SaveExpense(ctx context.Context, expense models.Expense) error
}

type ExpensesProvider interface {
	ListExpenses(ctx context.Context, category string, month, year int64) ([]models.Expense, int, error)
}

type CategoriesProvider interface {
	ListCategories(ctx context.Context) ([]models.Category, error)
}

type CategoriesReportProvider interface {
	ListCategoriesReport(ctx context.Context, filter string, month int, year int) ([]models.CategoryReport, error)
	Total(filter string, month int, year int) (int64, error)
}

func New(
	log *slog.Logger,
	expenseSaver ExpensesSaver, // TODO: use mock
	expensesProvider ExpensesProvider, // TODO: use mock
	categoriesReportProvider CategoriesReportProvider,
	categoriesProvider CategoriesProvider, // TODO: use mock
) *Finances {
	return &Finances{
		expenseSaver:             expenseSaver,
		expensesProvider:         expensesProvider,
		categoriesReportProvider: categoriesReportProvider,
		categoriesProvider:       categoriesProvider,
		log:                      log,
	}
}

func (f *Finances) Expense(
	ctx context.Context,
	Description string,
	Amount int64, // in cents
	Date string, // YYYY-MM-DD
	Category string, // "food", "groceries", "transport", "misc"
) (err error) {
	expense := models.Expense{
		Description: Description,
		Amount:      Amount,
		Date:        Date,
		Category:    Category,
	}

	err = f.expenseSaver.SaveExpense(ctx, expense)
	if err != nil {
		f.log.Error(err.Error())
		return err
	}

	return nil
}

func (f *Finances) ExpensesList(
	ctx context.Context, category string, month int64, year int64,
) (list []financesgrpc.Expense, total int64, err error) {
	l, t, err := f.expensesProvider.ListExpenses(ctx, category, month, year)

	if err != nil {
		f.log.Error(err.Error())
		return nil, 0, err
	}

	for _, e := range l {
		list = append(list, financesgrpc.Expense{
			ID:          int64(e.ID),
			Description: e.Description,
			Amount:      e.Amount,
			Date:        e.Date,
			Category:    e.Category,
			Color:       e.Color,
		})
	}
	return list, int64(t), nil
}

func (f *Finances) CategoriesList(ctx context.Context) ([]financesgrpc.Category, error) {
	categories, err := f.categoriesProvider.ListCategories(ctx)
	if err != nil {
		f.log.Error(err.Error())
	}

	var list []financesgrpc.Category
	for _, c := range categories {
		list = append(list, financesgrpc.Category{
			Name: c.Name,
			ID:   c.ID,
		})
	}

	return list, nil

}

func (f *Finances) CreateCategory(ctx context.Context, _ string) (string, error) { return "", nil }

func (f *Finances) Report(ctx context.Context, rf financesgrpc.ReportFilter, month int, year int) (int64, []financesgrpc.CategoryReport, error) {
	cts, err := f.categoriesReportProvider.ListCategoriesReport(ctx, rf.String(), month, year)

	if err != nil {
		f.log.Error(err.Error())
		return 0, nil, err
	}

	total, err := f.categoriesReportProvider.Total(rf.String(), month, year)
	if err != nil {
		f.log.Error(err.Error())
		return 0, nil, err
	}

	var cts2 []financesgrpc.CategoryReport

	for _, ct := range cts {
		cts2 = append(cts2, financesgrpc.CategoryReport{
			Category: ct.Name,
			Amount:   ct.Amount,
			Color:    ct.Color,
		})
	}
	return total, cts2, nil
}
