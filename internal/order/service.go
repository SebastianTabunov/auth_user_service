package order

type Service interface {
	GetOrder(orderID, userID int) (*Order, error)
	CreateOrder(userID int, title, description string, price float64) (*Order, error)
	GetUserOrders(userID int) ([]Order, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetOrder(orderID, userID int) (*Order, error) {
	return s.repo.GetOrder(orderID, userID)
}

func (s *service) CreateOrder(userID int, title, description string, price float64) (*Order, error) {
	order := &Order{
		UserID:      userID,
		Title:       title,
		Description: description,
		Price:       price,
		Status:      "pending",
	}

	id, err := s.repo.CreateOrder(order)
	if err != nil {
		return nil, err
	}

	order.ID = id
	return order, nil
}

func (s *service) GetUserOrders(userID int) ([]Order, error) {
	return s.repo.GetUserOrders(userID)
}
