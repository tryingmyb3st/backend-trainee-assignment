package teams

import (
	"errors"
	"sync"
)

var (
	ERROR_TEAM_EXISTS = errors.New("team_name already exists")
	ERROR_NOT_FOUND   = errors.New("resource not found")
)

type TeamsRepository struct {
	Teams map[string]Team
	Mtx   sync.RWMutex
}

func NewTeamsRepo() *TeamsRepository {
	return &TeamsRepository{
		Teams: make(map[string]Team),
		Mtx:   sync.RWMutex{},
	}
}

func (teamRepo *TeamsRepository) CreateNewTeam(newTeam Team) (Team, error) {
	teamRepo.Mtx.Lock()
	defer teamRepo.Mtx.Unlock()

	if _, exist := teamRepo.Teams[newTeam.TeamName]; exist {
		return Team{}, ERROR_TEAM_EXISTS
	}

	teamRepo.Teams[newTeam.TeamName] = newTeam
	return teamRepo.Teams[newTeam.TeamName], nil
}

func (teamRepo *TeamsRepository) GetTeamByName(name string) (Team, error) {
	teamRepo.Mtx.RLock()
	defer teamRepo.Mtx.RUnlock()

	if _, exist := teamRepo.Teams[name]; !exist {
		return Team{}, ERROR_NOT_FOUND
	}

	return teamRepo.Teams[name], nil
}

func (teamRepo *TeamsRepository) ChangeTeamMemberStatus(teamName string, userId string, status bool) {
	teamRepo.Mtx.Lock()
	defer teamRepo.Mtx.Unlock()

	team := teamRepo.Teams[teamName]
	for i, member := range team.Members {
		if member.Id == userId {
			team.Members[i].IsActive = status
		}
	}
	teamRepo.Teams[teamName] = team
}
