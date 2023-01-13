package model

import (
	"encoding/json"
	"github.com/google/uuid"
	"time"
)

type Job struct {
	Id        uuid.UUID `gorm:"type:uuid;primaryKey;index;not null;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time `gorm:"autoCreateTime;not null;default:now();" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;not null;default:now();" json:"updated_at"`

	UserId          int64           `gorm:"type:int64;index;not null;" json:"user_id"`
	State           string          `gorm:"type:varchar(100);not null;" json:"state"`
	Comment         string          `gorm:"type:varchar(250);" json:"comment"`
	Retry           int16           `gorm:"type:int;not null;default:0" json:"retry"`
	Result          json.RawMessage `json:"result"`
	Params          json.RawMessage `json:"params"`
	ReportType      string          `gorm:"type:varchar(100);not null;" json:"report_type"`
	ReportGenerated bool            `gorm:"type:boolean;not null;default:false" json:"report_generated"`
}
