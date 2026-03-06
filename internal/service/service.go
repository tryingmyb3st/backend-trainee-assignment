package service

import (
	"backend-task/internal/models"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
)

type PullReqRepository interface {
	GetPullRequest(id string) (*models.PullRequest, error)
	GetPullRequestsByReviewer(id string) ([]models.PullRequest, error)
	AddPullRequest(newPR models.PullRequest) error
	UpdatePullRequest(id string, pr models.PullRequest) error
	DeletePullRequst(id string) error
}

type UsersRepository interface {
	GetUser(id string) (*models.User, error)
	GetUsersByTeam(teamName string) (*models.Team, error)
	GetActiveUserFromTeam(teamName string) (*models.User, error)
	AddNewUser(newUser models.User, teamName string) error
	UpdateUser(id string, data models.User) error
	DeleteUser(id string) error
}

type Service struct {
	prRepo    PullReqRepository
	usersRepo UsersRepository
}

func NewService(PRrepo PullReqRepository, UsersRepo UsersRepository) Service {
	return Service{
		prRepo:    PRrepo,
		usersRepo: UsersRepo,
	}
}

func (s *Service) CreateNewTeam(team models.Team) error {
	if users, _ := s.usersRepo.GetUsersByTeam(team.TeamName); len(users.Members) != 0 {
		return models.ERROR_TEAM_EXISTS
	}

	for _, member := range team.Members {
		err := s.usersRepo.AddNewUser(member, team.TeamName)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) GetTeamByName(teamName string) (*models.Team, error) {
	team, err := s.usersRepo.GetUsersByTeam(teamName)
	if err != nil {
		return nil, err
	}

	if len(team.Members) == 0 {
		return nil, models.ERROR_NOT_FOUND
	}

	return team, nil
}

func (s *Service) ChangeUserStatus(user models.UserWithStatus) error {
	// а надо освобождать от пуллреквеста ?

	data, err := s.usersRepo.GetUser(user.UserId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.ERROR_NOT_FOUND
		}
		return err
	}

	data.IsActive = user.IsActive
	err = s.usersRepo.UpdateUser(user.UserId, *data)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) GetUserByID(id string) (*models.User, error) {
	user, err := s.usersRepo.GetUser(id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Service) CreateNewPR(pr models.PullRequest) error {
	// а если у автора уже есть пуллреквест ?

	if _, err := s.usersRepo.GetUser(pr.AuthorId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.ERROR_NOT_FOUND
		}
		return err
	}

	if exists, _ := s.prRepo.GetPullRequest(pr.Id); exists != nil {
		return models.ERROR_PR_EXISTS
	}

	pr.Status = "OPEN"

	// а если параллельно ?
	author, err := s.usersRepo.GetUser(pr.AuthorId)
	if err != nil {
		return err
	}

	err = s.ChangeUserStatus(models.UserWithStatus{
		UserId:   author.Id,
		IsActive: false,
	})
	if err != nil {
		return err
	}

	pr.AssignedReviewers = make([]string, 0, 2)
	for range 2 {
		reviewer, err := s.assignReviewer(author.TeamName)
		if err != nil && !errors.Is(err, models.ERROR_NOT_FOUND) {
			return err
		} else if err == nil {
			pr.AssignedReviewers = append(pr.AssignedReviewers, reviewer.Id)
		}
	}

	if err := s.prRepo.AddPullRequest(pr); err != nil {
		return err
	}

	return nil
}

func (s *Service) assignReviewer(teamName string) (models.User, error) {
	// mutex ?

	user, err := s.usersRepo.GetActiveUserFromTeam(teamName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, models.ERROR_NOT_FOUND
		}
		return models.User{}, err
	}

	user.IsActive = false
	if err = s.usersRepo.UpdateUser(user.Id, *user); err != nil {
		return models.User{}, err
	}

	return *user, nil
}

func (s *Service) GetPRByID(id string) (*models.PullRequest, error) {
	pr, err := s.prRepo.GetPullRequest(id)
	return pr, err
}

func (s *Service) GetUserPrs(id string) ([]models.PullRequest, error) {
	prs, err := s.prRepo.GetPullRequestsByReviewer(id)
	if err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("Error with db: %w", err)
	}
	return prs, nil
}

func (s *Service) MergePullRequest(id string) error {
	pr, err := s.prRepo.GetPullRequest(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.ERROR_NOT_FOUND
		}
		return err
	}

	if pr.Status == "MERGED" {
		return nil
	}

	pr.Status = "MERGED"
	timeNow := time.Now()
	pr.MergedAt = &timeNow

	err = s.ChangeUserStatus(models.UserWithStatus{
		UserId:   pr.AuthorId,
		IsActive: true,
	})
	if err != nil {
		return err
	}
	for _, reviewerId := range pr.AssignedReviewers {
		err = s.ChangeUserStatus(models.UserWithStatus{
			UserId:   reviewerId,
			IsActive: true,
		})
		if err != nil {
			return err
		}
	}

	err = s.prRepo.UpdatePullRequest(pr.Id, *pr)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (s *Service) ReassignNewReviewer(pr models.PullRequest) (*models.User, error) {
	prInBase, err := s.prRepo.GetPullRequest(pr.Id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ERROR_NOT_FOUND
		}
		return nil, err
	}

	if strings.EqualFold(prInBase.Status, "MERGED") {
		return nil, models.ERROR_PR_MERGED
	}

	if !slices.Contains(prInBase.AssignedReviewers, pr.OldReviewerId) {
		return nil, models.ERROR_NOT_ASSIGNED
	}

	author, err := s.usersRepo.GetUser(prInBase.AuthorId)
	if err != nil {
		return nil, err
	}

	replacingUser, err := s.usersRepo.GetActiveUserFromTeam(author.TeamName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ERROR_NO_CANDIDATE
		}
		return nil, err
	}

	err = s.ChangeUserStatus(models.UserWithStatus{
		UserId:   pr.OldReviewerId,
		IsActive: true,
	})
	if err != nil {
		return nil, err
	}

	err = s.ChangeUserStatus(models.UserWithStatus{
		UserId:   replacingUser.Id,
		IsActive: false,
	})
	if err != nil {
		return nil, err
	}

	prInBase.AssignedReviewers = slices.DeleteFunc(prInBase.AssignedReviewers, func(reviewer string) bool {
		return reviewer == pr.OldReviewerId
	})
	prInBase.AssignedReviewers = append(prInBase.AssignedReviewers, replacingUser.Id)

	if err := s.prRepo.UpdatePullRequest(prInBase.Id, *prInBase); err != nil {
		return nil, err
	}

	return replacingUser, nil
}
