package dto

import (
	"GAMERS-BE/internal/contest/application/port"
	"time"
)

type SenderResponse struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Tag      string `json:"tag"`
	Avatar   string `json:"avatar,omitempty"`
}

type ApplicationResponse struct {
	UserID      int64                    `json:"user_id"`
	ContestID   int64                    `json:"contest_id"`
	Status      port.ApplicationStatus   `json:"status"`
	RequestedAt time.Time                `json:"requested_at"`
	ProcessedAt *time.Time               `json:"processed_at,omitempty"`
	ProcessedBy *int64                   `json:"processed_by,omitempty"`
	Sender      *SenderResponse          `json:"sender,omitempty"`
}

func ToApplicationResponse(app *port.ContestApplication) *ApplicationResponse {
	var sender *SenderResponse
	if app.Sender != nil {
		sender = &SenderResponse{
			UserID:   app.Sender.UserID,
			Username: app.Sender.Username,
			Tag:      app.Sender.Tag,
			Avatar:   app.Sender.Avatar,
		}
	}

	return &ApplicationResponse{
		UserID:      app.UserID,
		ContestID:   app.ContestID,
		Status:      app.Status,
		RequestedAt: app.RequestedAt,
		ProcessedAt: app.ProcessedAt,
		ProcessedBy: app.ProcessedBy,
		Sender:      sender,
	}
}

func ToApplicationResponses(apps []*port.ContestApplication) []*ApplicationResponse {
	responses := make([]*ApplicationResponse, len(apps))
	for i, app := range apps {
		responses[i] = ToApplicationResponse(app)
	}
	return responses
}
