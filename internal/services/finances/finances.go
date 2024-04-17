package finances

import (
	"context"
	financesgrpc "github.com/kochnevns/finances-backend/internal/grpc/finances"
	"github.com/kochnevns/finances-backend/internal/models"
	"log/slog"
)

type Finances struct {
	log              *slog.Logger
	expenseSaver     ExpensesSaver
	expensesProvider ExpensesProvider
}

//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=URLSaver
type ExpensesSaver interface {
	SaveExpense(ctx context.Context, expense models.Expense) error
}

type ExpensesProvider interface {
	ListExpenses(ctx context.Context) ([]models.Expense, error)
}

func New(
	log *slog.Logger,
	expenseSaver ExpensesSaver, // TODO: use mock
	expensesProvider ExpensesProvider, // TODO: use mock
) *Finances {
	return &Finances{
		expenseSaver:     expenseSaver,
		expensesProvider: expensesProvider,
		log:              log,
	}
}

// Login checks if user with given credentials exists in the system and returns access token.
//
// If user exists, but password is incorrect, returns error.
// If user doesn't exist, returns error.

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
	ctx context.Context,
) (list []financesgrpc.Expense, err error) {
	l, err := f.expensesProvider.ListExpenses(ctx)

	if err != nil {
		f.log.Error(err.Error())
		return nil, err
	}

	for _, e := range l {
		list = append(list, financesgrpc.Expense{
			Description: e.Description,
			Amount:      e.Amount,
			Date:        e.Date,
			Category:    e.Category,
		})
	}
	return list, nil // TODO: return error
}

func (f *Finances) CategoriesList(ctx context.Context, _ string) (string, error) { return "", nil }

func (f *Finances) CreateCategory(ctx context.Context, _ string) (string, error) { return "", nil }
