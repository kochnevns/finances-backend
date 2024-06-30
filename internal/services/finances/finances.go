package finances

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	financesgrpc "github.com/kochnevns/finances-backend/internal/grpc/finances"
	"github.com/kochnevns/finances-backend/internal/imcache"
	"github.com/kochnevns/finances-backend/internal/models"
)

type Finances struct {
	log                      *slog.Logger
	expenseSaver             ExpensesSaver
	expenseUpdater           ExpenseUpdater
	expensesProvider         ExpensesProvider
	categoriesReportProvider CategoriesReportProvider
	categoriesProvider       CategoriesProvider
	cache                    *imcache.IMCache
}

//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=URLSaver
type ExpensesSaver interface {
	SaveExpense(ctx context.Context, expense models.Expense) error
}

type ExpenseUpdater interface {
	UpdateExpense(ctx context.Context, expense models.Expense) error
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
	MedianAndMiddle(ctx context.Context, month int, year int) (int64, int64, error)
}

func New(
	log *slog.Logger,
	expenseSaver ExpensesSaver, // TODO: use mock
	expensesUpdater ExpenseUpdater,
	expensesProvider ExpensesProvider, // TODO: use mock
	categoriesReportProvider CategoriesReportProvider,
	categoriesProvider CategoriesProvider, // TODO: use mock
	cache *imcache.IMCache,
) *Finances {
	return &Finances{
		expenseSaver:             expenseSaver,
		expenseUpdater:           expensesUpdater,
		expensesProvider:         expensesProvider,
		categoriesReportProvider: categoriesReportProvider,
		categoriesProvider:       categoriesProvider,
		log:                      log,
		cache:                    cache,
	}
}

func (f *Finances) Expense(
	ctx context.Context,
	Description string,
	Amount int64, // in cents
	Date string, // YYYY-MM-DD
	Category string, // "food", "groceries", "transport", "misc"
	Id int64,
) (err error) {
	expense := models.Expense{
		ID:          Id,
		Description: Description,
		Amount:      Amount,
		Date:        Date,
		Category:    Category,
	}

	f.cache.Flush()

	if Id == 0 {
		err := f.expenseSaver.SaveExpense(ctx, expense)
		if err != nil {
			f.log.Error(err.Error())
			return err
		}
	} else {
		if err := f.expenseUpdater.UpdateExpense(ctx, expense); err != nil {
			f.log.Error(err.Error())
			return err
		}
	}

	return nil
}

func (f *Finances) ExpensesList(
	ctx context.Context, category string, month int64, year int64,
) (list []financesgrpc.Expense, total int64, err error) {

	cacheKeyList := fmt.Sprintf("list;%s;%d;%d", category, month, year)
	cacheKeyTotal := fmt.Sprintf("total;%s;%d;%d", category, month, year)

	x, foundList := f.cache.Get(cacheKeyList)
	tot, foundTotal := f.cache.Get(cacheKeyTotal)

	if foundList && foundTotal {
		f.log.Info("Cache hit")
		l := x.(*[]financesgrpc.Expense)
		totInt := tot.(*int)

		return *l, int64(*totInt), nil
	}

	f.log.Info("Cache miss")

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

	f.cache.Set(cacheKeyList, &list, time.Hour)
	f.cache.Set(cacheKeyTotal, &t, time.Hour)

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

func (f *Finances) Report(ctx context.Context, rf financesgrpc.ReportFilter, month int, year int) (int64, int64, int64, []financesgrpc.CategoryReport, error) {
	cts, err := f.categoriesReportProvider.ListCategoriesReport(ctx, rf.String(), month, year)

	if err != nil {
		f.log.Error(err.Error())
		return 0, 0, 0, nil, err
	}

	total, err := f.categoriesReportProvider.Total(rf.String(), month, year)
	if err != nil {
		f.log.Error(err.Error())
		return 0, 0, 0, nil, err
	}

	var cts2 []financesgrpc.CategoryReport

	for _, ct := range cts {
		cts2 = append(cts2, financesgrpc.CategoryReport{
			Category: ct.Name,
			Amount:   ct.Amount,
			Color:    ct.Color,
		})
	}

	middle, median, err := f.categoriesReportProvider.MedianAndMiddle(ctx, month, year)

	if err != nil {
		return 0, 0, 0, nil, err
	}

	return total, middle, median, cts2, nil
}
