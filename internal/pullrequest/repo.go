package pullrequest

import (
	"backend-task/internal/teams"
	"backend-task/internal/user"
	"errors"
	"slices"
	"sync"
	"time"
)

var (
	ERROR_PR_EXISTS    = errors.New("PR id already exists")
	ERROR_PR_NOT_FOUND = errors.New("resource not found")
	ERROR_PR_MERGED    = errors.New("cannot reassign on merged PR")
)

type PRrepository struct {
	PullRequests map[string]PullRequest
	Mtx          sync.RWMutex
}

func NewPrRepo() *PRrepository {
	return &PRrepository{
		PullRequests: make(map[string]PullRequest),
		Mtx:          sync.RWMutex{},
	}
}

func (prRepo *PRrepository) CreateNewPR(newPR NewPullRequest, userRepo *user.UsersRepository, teamsRepo *teams.TeamsRepository) (PullRequest, error) {
	prRepo.Mtx.Lock()
	defer prRepo.Mtx.Unlock()

	if _, exist := prRepo.PullRequests[newPR.PullRequestId]; exist {
		return PullRequest{}, ERROR_PR_EXISTS
	}

	authorTeamName := userRepo.Users[newPR.AuthorId].TeamName
	reviewersToAssign := make([]string, 0, 2)
	for _, user := range userRepo.Users {
		if len(reviewersToAssign) == 2 {
			break
		}
		if user.TeamName == authorTeamName && user.Id != newPR.AuthorId && user.IsActive {
			reviewersToAssign = append(reviewersToAssign, user.Id)
			userRepo.ChangeUserStatus(user.Id, false)
			teamsRepo.ChangeTeamMemberStatus(authorTeamName, user.Id, false)
		}
	}
	userRepo.ChangeUserStatus(newPR.AuthorId, false)
	teamsRepo.ChangeTeamMemberStatus(authorTeamName, newPR.AuthorId, false)

	pr := PullRequest{
		Id:                newPR.PullRequestId,
		Name:              newPR.PullRequstName,
		AuthorId:          newPR.AuthorId,
		Status:            OPEN,
		AssignedReviewers: reviewersToAssign,
	}
	prRepo.PullRequests[newPR.PullRequestId] = pr
	return pr, nil
}

func (prRepo *PRrepository) MergePR(pr NewPullRequest, userRepo *user.UsersRepository, teamsRepo *teams.TeamsRepository) (PullRequest, error) {
	prRepo.Mtx.Lock()
	defer prRepo.Mtx.Unlock()

	if _, exist := prRepo.PullRequests[pr.PullRequestId]; !exist {
		return PullRequest{}, ERROR_PR_NOT_FOUND
	}

	if prRepo.PullRequests[pr.PullRequestId].Status == MERGED {
		return prRepo.PullRequests[pr.PullRequestId], nil
	}

	prToMerge := prRepo.PullRequests[pr.PullRequestId]
	prToMerge.Status = MERGED
	prToMerge.MergedAt = time.Now()
	prRepo.PullRequests[pr.PullRequestId] = prToMerge

	for _, userId := range prToMerge.AssignedReviewers {
		userRepo.ChangeUserStatus(userId, true)
		teamsRepo.ChangeTeamMemberStatus(userRepo.Users[userId].TeamName, userId, true)
	}
	userRepo.ChangeUserStatus(prToMerge.AuthorId, true)
	teamsRepo.ChangeTeamMemberStatus(userRepo.Users[prToMerge.AuthorId].TeamName, prToMerge.AuthorId, true)
	return prToMerge, nil
}

func (prRepo *PRrepository) ReassignUser(pr NewPullRequest, userRepo *user.UsersRepository, teamsRepo *teams.TeamsRepository) (PullRequest, string, error) {
	prRepo.Mtx.Lock()
	defer prRepo.Mtx.Unlock()

	if _, exist := prRepo.PullRequests[pr.PullRequestId]; !exist {
		return PullRequest{}, "", ERROR_PR_NOT_FOUND
	}

	if val := prRepo.PullRequests[pr.PullRequestId]; val.Status == MERGED {
		return PullRequest{}, "", ERROR_PR_MERGED
	}

	prToModify := prRepo.PullRequests[pr.PullRequestId]
	currentReviewers := prToModify.AssignedReviewers
	modifiedtReviewers := slices.DeleteFunc(currentReviewers, func(userId string) bool {
		return userId == pr.OldReviewerId
	})

	teamName := userRepo.Users[pr.OldReviewerId].TeamName
	userRepo.ChangeUserStatus(pr.OldReviewerId, true)
	teamsRepo.ChangeTeamMemberStatus(teamName, pr.OldReviewerId, true)

	var replacedBy string
	for _, member := range teamsRepo.Teams[teamName].Members {
		if !slices.Contains(modifiedtReviewers, member.Id) && member.Id != pr.OldReviewerId && member.IsActive {
			modifiedtReviewers = append(modifiedtReviewers, member.Id)
			replacedBy = member.Id
			userRepo.ChangeUserStatus(member.Id, false)
			teamsRepo.ChangeTeamMemberStatus(teamName, member.Id, false)
			break
		}
	}

	prToModify.AssignedReviewers = modifiedtReviewers
	prRepo.PullRequests[pr.PullRequestId] = prToModify
	return prToModify, replacedBy, nil
}

func (PRrepo *PRrepository) GetUserPRs(userId string, userRepo *user.UsersRepository) ([]PullRequest, error) {
	userRepo.Mtx.RLock()
	defer userRepo.Mtx.RUnlock()

	if _, exist := userRepo.Users[userId]; !exist {
		return []PullRequest{}, teams.ERROR_NOT_FOUND
	}

	userTasks := make([]PullRequest, 0)
	for _, pr := range PRrepo.PullRequests {
		if slices.Contains(pr.AssignedReviewers, userId) || pr.AuthorId == userId {
			pr.AssignedReviewers = make([]string, 0)
			userTasks = append(userTasks, pr)
		}
	}

	return userTasks, nil
}
