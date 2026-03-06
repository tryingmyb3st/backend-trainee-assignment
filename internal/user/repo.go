package user

import (
	"backend-task/internal/models"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type UsersRepository struct {
	db  *pgx.Conn
	ctx context.Context
}

func NewRepo(conn *pgx.Conn, context context.Context) *UsersRepository {
	return &UsersRepository{
		db:  conn,
		ctx: context,
	}
}

func (repo *UsersRepository) GetUser(id string) (*models.User, error) {
	user := models.User{}

	SQLQuery := `SELECT * FROM users WHERE id=$1`

	row := repo.db.QueryRow(repo.ctx, SQLQuery, id)

	err := row.Scan(&user.Id, &user.Username, &user.IsActive, &user.TeamName)
	if err != nil {
		return nil, fmt.Errorf("Error in row.Scan: %w", err)
	}

	return &user, nil
}

func (repo *UsersRepository) GetActiveUserFromTeam(teamName string) (*models.User, error) {
	user := models.User{}

	SQLQuery := `SELECT * FROM users WHERE team=$1 AND is_active=$2`

	row := repo.db.QueryRow(repo.ctx, SQLQuery, teamName, true)

	err := row.Scan(&user.Id, &user.Username, &user.IsActive, &user.TeamName)
	if err != nil {
		return nil, fmt.Errorf("Error in row.Scan: %w", err)
	}

	return &user, nil
}

func (repo *UsersRepository) AddNewUser(newUser models.User, teamName string) error {
	SQLQuery := `INSERT INTO users(id, username, is_active, team)
	VALUES($1, $2, $3, $4);`
	_, err := repo.db.Exec(repo.ctx, SQLQuery, newUser.Id, newUser.Username, newUser.IsActive, teamName)
	return err
}

func (repo *UsersRepository) UpdateUser(id string, user models.User) error {
	SQLQuery := `UPDATE users 
	SET username=$2, is_active=$3, team=$4
	WHERE id=$1;`
	_, err := repo.db.Exec(repo.ctx, SQLQuery, user.Id, user.Username, user.IsActive, user.TeamName)
	return err
}

func (repo *UsersRepository) DeleteUser(id string) error {
	SQLQuery := `DELETE FROM users
	WHERE id=$1`
	_, err := repo.db.Exec(repo.ctx, SQLQuery, id)
	return err
}

func (repo *UsersRepository) GetUsersByTeam(teamName string) (*models.Team, error) {
	team := models.Team{
		TeamName: teamName,
		Members:  make([]models.User, 0),
	}

	SQLQuery := `SELECT id, username, is_active FROM users
	WHERE team=$1`
	rows, err := repo.db.Query(repo.ctx, SQLQuery, teamName)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		member := models.User{}
		err := rows.Scan(&member.Id, &member.Username, &member.IsActive)
		if err != nil {
			return nil, err
		}
		team.Members = append(team.Members, member)
	}

	return &team, nil
}
