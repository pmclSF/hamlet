package services

type UserService struct{}

func NewUserService() *UserService {
	return &UserService{}
}

func (s *UserService) FindByID(id string) string {
	return "user:" + id
}

func (s *UserService) Create(name string) string {
	return "created:" + name
}
