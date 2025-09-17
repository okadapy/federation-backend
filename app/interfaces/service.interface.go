package interfaces

type Service[Model interface{}] interface {
	Create(dto interface{}) error
	Get(id uint) (Model, error)
	Update(id uint, dto interface{}) error
	Delete(id uint) error
}
