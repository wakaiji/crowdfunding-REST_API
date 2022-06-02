package campaign

import (
	"bwastartup/user"
	"time"
)

type Campaign struct {
	ID               int       `json:"id"`
	UserID           int       `json:"user_id"`
	Name             string    `json:"name"`
	ShortDescription string    `json:"short_description"`
	Description      string    `json:"description"`
	Perks            string    `json:"perks"`
	BackerCount      int       `json:"backer_count"`
	GoalAmount       int       `json:"goal_amount"`
	CurrentAmount    int       `json:"current_amount"`
	Slug             string    `json:"slug"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	CampaignImages   []CampaignImage
	User             user.User
}

type CampaignImage struct {
	ID         int       `json:"id"`
	CampaignID int       `json:"campaign_id"`
	FileName   string    `json:"file_name"`
	IsPrimary  int       `json:"is_primary"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
