package cloud

import "time"

type Actor struct {
	Id                string    `json:"id"`
	Handle            string    `json:"handle"`
	DisplayName       string    `json:"display_name"`
	AvatarUrl         string    `json:"avatar_url"`
	Status            string    `json:"status"`
	PreviewAccessMode string    `json:"preview_access_mode"`
	VersionId         float64   `json:"version_id"`
	CreatedAt         time.Time `json:"created_at"`
	UpdateAt          time.Time `json:"updated_at"`
}
