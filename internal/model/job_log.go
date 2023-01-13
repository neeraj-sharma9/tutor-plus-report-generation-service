package model

import (
	"encoding/json"
	"github.com/google/uuid"
	"time"
)

type JobLog struct {
	Id        uuid.UUID `gorm:"type:uuid;primaryKey;index;not null;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time `gorm:"autoCreateTime;not null;default:now();" json:"created_at"`

	JobId           uuid.UUID       `gorm:"type:uuid;index;not null;default:null;" json:"job_id" sql:"type:uuid REFERENCES job(id)"`
	Job             *Job            `gorm:"foreignKey:JobId;not null;default:null;" json:"job,omitempty"`
	State           string          `gorm:"type:varchar(100);not null;" json:"state"`
	Comment         string          `gorm:"type:varchar(250);" json:"comment"`
	Retry           int32           `gorm:"type:int;not null;default:0" json:"retry"`
	Result          json.RawMessage `json:"result"`
	Params          json.RawMessage `json:"params"`
	ReportGenerated bool            `gorm:"type:boolean;not null;default:false" json:"report_generated"`
}
