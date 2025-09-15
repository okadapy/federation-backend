package team

import "federation-backend/app/db/models"

type Service struct {
}

func (s Service) Create(dto models.Team) error {
	//TODO implement me
	panic("implement me")
}

func (s Service) Get(id uint) (models.Team, error) {
	//TODO implement me
	panic("implement me")
}

func (s Service) Update(id uint, dto models.Team) error {
	//TODO implement me
	panic("implement me")
}

func (s Service) Delete(id uint) error {
	//TODO implement me
	panic("implement me")
}
