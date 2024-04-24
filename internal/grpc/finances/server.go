package financesgrpc

import (
	"context"

	financesgrpcsrv "github.com/kochnevns/finances-protos/finances"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Expense struct {
	ID          int64
	Description string
	Amount      int64  // in cents
	Date        string // YYYY-MM-DD
	Category    string // "food", "groceries", "transport", "misc"
}

type ReportFilter int32

const (
	Week  ReportFilter = 0
	Month ReportFilter = 1
	Year  ReportFilter = 2
)

type CategoryReport struct {
	Category string
	Amount   int64
	Percent  float64
}

type Category struct {
	ID   int64
	Name string // "food", "groceries", "transport", "misc"
}

type Finances interface {
	Expense(
		ctx context.Context,
		Description string,
		Amount int64, // in cents
		Date string, // YYYY-MM-DD
		Category string, // "food", "groceries", "transport", "misc"
	) (err error)
	ExpensesList(
		ctx context.Context,
	) (list []Expense, err error)

	CreateCategory(context.Context, string) (string, error)
	CategoriesList(context.Context) ([]Category, error)
	Report(context.Context, ReportFilter) (int64, []CategoryReport, error)
}

type serverAPI struct {
	financesgrpcsrv.UnimplementedFinancesServer
	finances Finances
}

func Register(gRPCServer *grpc.Server, finances Finances) {
	financesgrpcsrv.RegisterFinancesServer(gRPCServer, &serverAPI{
		finances: finances,
	})
}

func (s *serverAPI) Report(ctx context.Context, in *financesgrpcsrv.ReportRequest) (*financesgrpcsrv.ReportResponse, error) {
	total, report, err := s.finances.Report(ctx, ReportFilter(in.Type))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var categories []*financesgrpcsrv.ReportCategory // TODO: make this an array

	for _, category := range report {
		categories = append(categories, &financesgrpcsrv.ReportCategory{
			Name:    category.Category,
			Amount:  category.Amount,
			Percent: category.Amount * 100 / total,
		})
	}

	rsp := &financesgrpcsrv.ReportResponse{
		Total:      total,
		Categories: categories,
	}

	return rsp, nil
}

func (s *serverAPI) Expense(
	ctx context.Context,
	in *financesgrpcsrv.ExpenseRequest,
) (*financesgrpcsrv.ExpenseResponse, error) {

	err := s.finances.Expense(
		ctx,
		in.Description,
		in.Amount,
		in.Date,
		in.Category,
	)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &financesgrpcsrv.ExpenseResponse{}, nil
}

func (s *serverAPI) ExpensesList(ctx context.Context, _ *financesgrpcsrv.ExpensesListRequest) (*financesgrpcsrv.ExpensesListResponse, error) {
	list, err := s.finances.ExpensesList(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	rsp := &financesgrpcsrv.ExpensesListResponse{}
	var respList []*financesgrpcsrv.Expense

	for _, expense := range list {
		respList = append(respList, &financesgrpcsrv.Expense{
			Amount:      expense.Amount,
			Date:        expense.Date,
			Category:    expense.Category,
			Description: expense.Description,
		})
	}

	rsp.Expenses = respList

	return rsp, nil
}

//	func (s *serverAPI) CreateCategory(ctx context.Context, category string) (string, error) {
//		return category, nil
//	}
func (s *serverAPI) CategoriesList(ctx context.Context, _ *financesgrpcsrv.CategoriesListRequest) (*financesgrpcsrv.CategoriesListResponse, error) {
	categories, err := s.finances.CategoriesList(ctx)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	rsp := &financesgrpcsrv.CategoriesListResponse{
		Categories: []*financesgrpcsrv.Category{},
	}
	for _, category := range categories {
		rsp.Categories = append(rsp.Categories, &financesgrpcsrv.Category{
			Name: category.Name,
		})
	}

	return rsp, nil
}
