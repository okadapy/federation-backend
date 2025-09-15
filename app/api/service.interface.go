package api

type Service[Dto interface{}] interface {
	Create(dto Dto) error
	Get(id uint) (Dto, error)
	Update(id uint, dto Dto) error
	Delete(id uint) error
}
