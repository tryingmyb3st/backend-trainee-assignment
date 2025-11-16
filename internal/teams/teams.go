package teams

type TeamMember struct {
	Id       string `json:"user_id"`
	Name     string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type Team struct {
	TeamName string       `json:"team_name"`
	Members  []TeamMember `json:"members"`
}
