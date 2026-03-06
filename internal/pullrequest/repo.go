package pullreq

import (
	"backend-task/internal/models"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type PRrepository struct {
	db  *pgx.Conn
	ctx context.Context
}

func NewRepo(conn *pgx.Conn, context context.Context) *PRrepository {
	return &PRrepository{
		db:  conn,
		ctx: context,
	}
}

func (repo *PRrepository) GetPullRequest(id string) (*models.PullRequest, error) {

	pr := models.PullRequest{}
	SQLQuery := `SELECT * FROM pullrequests WHERE id=$1`

	row := repo.db.QueryRow(repo.ctx, SQLQuery, id)

	err := row.Scan(&pr.Id, &pr.Name, &pr.AuthorId, &pr.Status, &pr.AssignedReviewers, &pr.MergedAt)
	if err != nil {
		return nil, fmt.Errorf("Error in row.Scan: %w", err)
	}

	return &pr, nil
}

func (repo *PRrepository) AddPullRequest(newPR models.PullRequest) error {
	SQLQuery := `INSERT INTO pullrequests(id, name, author_id, status, reviewers_id, merged_at)
	VALUES($1, $2, $3, $4, $5, $6);`
	_, err := repo.db.Exec(repo.ctx, SQLQuery, newPR.Id, newPR.Name, newPR.AuthorId, newPR.Status, newPR.AssignedReviewers, newPR.MergedAt)
	return err
}

func (repo *PRrepository) UpdatePullRequest(id string, pr models.PullRequest) error {
	SQLQuery := `UPDATE pullrequests 
	SET name=$2, author_id=$3, status=$4, reviewers_id=$5, merged_at=$6
	WHERE id=$1;`
	_, err := repo.db.Exec(repo.ctx, SQLQuery, pr.Id, pr.Name, pr.AuthorId, pr.Status, pr.AssignedReviewers, pr.MergedAt)
	return err
}

func (repo *PRrepository) DeletePullRequst(id string) error {
	SQLQuery := `DELETE FROM pullrequests
	WHERE id=$1`
	_, err := repo.db.Exec(repo.ctx, SQLQuery, id)
	return err
}

func (repo *PRrepository) DeleteReviewerFromPR(prID string, oldReviewerId string) error {
	SQLQuery := `UPDATE pullrequests
	SET reviewers_id=ARRAY_REMOVE(reviewers_id, $1)
	WHERE id=$2`
	_, err := repo.db.Exec(repo.ctx, SQLQuery, oldReviewerId, prID)
	return err
}

func (repo *PRrepository) GetPullRequestsByReviewer(id string) ([]models.PullRequest, error) {
	prs := make([]models.PullRequest, 0)

	SQLQuery := `SELECT * FROM pullrequests 
	WHERE $1 = ANY(reviewers_id);`

	rows, err := repo.db.Query(repo.ctx, SQLQuery, id)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		pr := models.PullRequest{}
		err = rows.Scan(&pr.Id, &pr.Name, &pr.AuthorId, &pr.Status, &pr.AssignedReviewers, &pr.MergedAt)
		if err != nil {
			return nil, err
		}
		pr.AssignedReviewers = []string{}
		prs = append(prs, pr)
	}

	return prs, nil
}
