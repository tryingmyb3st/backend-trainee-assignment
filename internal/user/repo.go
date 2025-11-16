package user

import (
	"backend-task/internal/teams"
	"sync"
)

type UsersRepository struct {
	Users map[string]User
	Mtx   sync.RWMutex
}

func NewUsersRepo() *UsersRepository {
	return &UsersRepository{
		Users: make(map[string]User),
		Mtx:   sync.RWMutex{},
	}
}

func (userRepo *UsersRepository) AddNewUsers(teamName string, users []teams.TeamMember) {
	userRepo.Mtx.Lock()
	defer userRepo.Mtx.Unlock()

	for _, val := range users {
		userRepo.Users[val.Id] = User{
			Id:       val.Id,
			Username: val.Name,
			TeamName: teamName,
			IsActive: val.IsActive,
		}
	}
}

func (userRepo *UsersRepository) ChangeUserStatus(userId string, status bool) (User, error) {
	userRepo.Mtx.Lock()
	defer userRepo.Mtx.Unlock()

	if _, exist := userRepo.Users[userId]; !exist {
		return User{}, teams.ERROR_NOT_FOUND
	}

	user := userRepo.Users[userId]
	user.IsActive = status
	userRepo.Users[userId] = user

	return user, nil
}

func (userRepo *UsersRepository) IsUserExists(userId string) bool {
	userRepo.Mtx.RLock()
	defer userRepo.Mtx.RUnlock()

	_, exist := userRepo.Users[userId]
	return exist
}
