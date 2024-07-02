package financesgrpc

import (
	"context"
	"fmt"
	"time"

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
	Color       string
}

type ReportFilter string

const (
	Week  ReportFilter = "week"
	Month ReportFilter = "month"
	Year  ReportFilter = "year" // TODO: make this an enum?
)

func (f ReportFilter) String() string {
	return string(f)
}

type CategoryReport struct {
	Category string
	Color    string
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
		Id int64, // expense ID in the database, not the ID in the gRPC request
	) (err error)

	// ExpenseEdit(
	// 	ctx context.Context,
	// 	ID int64, // expense ID in the database, not the ID in the gRPC request
	// 	Description string,
	// 	Amount int64, // in cents
	// 	Date string, // YYYY-MM-DD
	// 	Category string, // "food", "groceries", "transport", "misc"
	// ) (err error)
	ExpensesList(
		ctx context.Context,
		category string,
		month int64,
		year int64,
	) (list []Expense, totalAmount int64, err error)

	CreateCategory(context.Context, string) (string, error)
	CategoriesList(context.Context) ([]Category, error)
	Report(context.Context, ReportFilter, int, int) (int64, int64, int64, []CategoryReport, error)
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

func (s *serverAPI) MassiveReport(ctx context.Context, in *financesgrpcsrv.MassiveReportRequest) (*financesgrpcsrv.MassiveReportResponse, error) {
	response := &financesgrpcsrv.MassiveReportResponse{
		Monthes: []*financesgrpcsrv.ReportResponse{},
	}

	nowMonth := int(time.Now().Month())
	nowYear := time.Now().Year()

	for i := nowMonth - 12; i <= nowMonth; i++ {
		month := i
		year := nowYear
		if i < 1 {
			year--
			month = 12 + i
		}

		total, middle, median, report, err := s.finances.Report(ctx, Month, month, year)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to get report: %v", err)
		}

		reportResponse := &financesgrpcsrv.ReportResponse{
			Total: total,
			Month: fmt.Sprintf("%02d.%04d", month, year),
		}

		var categories []*financesgrpcsrv.ReportCategory // TODO: make this an array

		for _, category := range report {
			categories = append(categories, &financesgrpcsrv.ReportCategory{
				Name:    category.Category,
				Amount:  category.Amount,
				Percent: category.Amount * 100 / total,
				Color:   category.Color,
			})
		}

		reportResponse.Categories = categories
		reportResponse.Average = middle
		reportResponse.Median = median

		response.Monthes = append(response.Monthes, reportResponse)
	}

	return response, nil
}
func (s *serverAPI) ExpenseEdit(ctx context.Context, in *financesgrpcsrv.ExpenseEditRequest) (*financesgrpcsrv.ExpenseResponse, error) {
	// err := s.finances.Expense(ctx, in.Description, in.Amount, in.Date, in.Category)
	// if err != nil {
	// 	return nil, status.Error(codes.Internal, err.Error())
	// }
	// return &financesgrpcsrv.ExpenseResponse{
	// 	Description: in.Description,
	// 	Amount:      in.Amount,
	// 	Date:        in.Date,
	// 	Category:    in.Category,
	// }, nil

	return nil, nil // TODO: implement this
}

func (s *serverAPI) Report(ctx context.Context, in *financesgrpcsrv.ReportRequest) (*financesgrpcsrv.ReportResponse, error) {
	nowMonth := int(time.Now().Month())
	nowYear := time.Now().Year()

	total, _, _, report, err := s.finances.Report(ctx, ReportFilter(fmt.Sprintf("%s", in.GetType())), nowMonth, nowYear)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var categories []*financesgrpcsrv.ReportCategory // TODO: make this an array

	for _, category := range report {
		categories = append(categories, &financesgrpcsrv.ReportCategory{
			Name:    category.Category,
			Amount:  category.Amount,
			Percent: category.Amount * 100 / total,
			Color:   category.Color,
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
		in.GetId(),
	)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &financesgrpcsrv.ExpenseResponse{}, nil
}

func (s *serverAPI) ExpensesList(ctx context.Context, req *financesgrpcsrv.ExpensesListRequest) (*financesgrpcsrv.ExpensesListResponse, error) {

	list, totalAmount, err := s.finances.ExpensesList(ctx, req.GetCategory(), req.GetMonth(), req.GetYear())

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	rsp := &financesgrpcsrv.ExpensesListResponse{}
	respList := make([]*financesgrpcsrv.Expense, 0, len(list))

	for _, expense := range list {
		respList = append(respList, &financesgrpcsrv.Expense{
			Id:          expense.ID,
			Amount:      expense.Amount,
			Date:        expense.Date,
			Category:    expense.Category,
			Description: expense.Description,
			Color:       expense.Color,
		})
	}

	rsp.Expenses = respList
	rsp.Total = totalAmount

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
