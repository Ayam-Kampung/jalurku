package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Model tabel pengguna
type User struct {
	ID        uuid.UUID      `gorm:"type:char(36);primaryKey" json:"id"`
	Name      string         `gorm:"type:varchar(255);not null" json:"name"`
	Email     string         `gorm:"type:varchar(255);unique;not null" json:"email"`
	Password  string         `gorm:"type:varchar(255);not null" json:"-"`
	Role      string         `gorm:"type:varchar(20);default:'user'" json:"role"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Reflections []Reflection `gorm:"foreignKey:UserID" json:"reflections,omitempty"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// Category represents the categories table (Dimensi)
type Category struct {
	ID          uuid.UUID      `gorm:"type:char(36);primaryKey" json:"id"`
	Name        string         `gorm:"type:varchar(255);not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	Weight      float64        `gorm:"type:decimal(5,2);not null;default:0" json:"weight"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Questions []Question `gorm:"foreignKey:CategoryID" json:"questions,omitempty"`
}

func (c *Category) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// Question represents the questions table
type Question struct {
	ID         uuid.UUID      `gorm:"type:char(36);primaryKey" json:"id"`
	CategoryID uuid.UUID      `gorm:"type:char(36);not null" json:"category_id"`
	Text       string         `gorm:"type:text;not null" json:"text"`
	Type       string         `gorm:"type:varchar(50);default:'scale'" json:"type"`
	Order      int            `gorm:"default:0" json:"order"`
	IsActive   bool           `gorm:"default:true" json:"is_active"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Category Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Options  []Option `gorm:"foreignKey:QuestionID" json:"options,omitempty"`
	Answers  []Answer `gorm:"foreignKey:QuestionID" json:"answers,omitempty"`
}

func (q *Question) BeforeCreate(tx *gorm.DB) error {
	if q.ID == uuid.Nil {
		q.ID = uuid.New()
	}
	return nil
}

// Option represents the options table for multiple choice questions
type Option struct {
	ID         uuid.UUID      `gorm:"type:char(36);primaryKey" json:"id"`
	QuestionID uuid.UUID      `gorm:"type:char(36);not null" json:"question_id"`
	Text       string         `gorm:"type:varchar(255);not null" json:"text"`
	Score      int            `gorm:"default:0" json:"score"`
	Order      int            `gorm:"default:0" json:"order"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Question Question `gorm:"foreignKey:QuestionID" json:"question,omitempty"`
}

func (o *Option) BeforeCreate(tx *gorm.DB) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	return nil
}

// Reflection represents the reflections table (Session refleksi)
type Reflection struct {
	ID          uuid.UUID      `gorm:"type:char(36);primaryKey" json:"id"`
	UserID      uuid.UUID      `gorm:"type:char(36);not null" json:"user_id"`
	Title       string         `gorm:"type:varchar(255)" json:"title"`
	TotalScore  float64        `gorm:"type:decimal(10,2);default:0" json:"total_score"`
	MaxScore    float64        `gorm:"type:decimal(10,2);default:0" json:"max_score"`
	Percentage  float64        `gorm:"type:decimal(5,2);default:0" json:"percentage"`
	Status      string         `gorm:"type:varchar(20);default:'draft'" json:"status"`
	CompletedAt *time.Time     `json:"completed_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	User            User              `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Answers         []Answer          `gorm:"foreignKey:ReflectionID" json:"answers,omitempty"`
	CategoryScores  []CategoryScore   `gorm:"foreignKey:ReflectionID" json:"category_scores,omitempty"`
	Recommendations []Recommendation  `gorm:"foreignKey:ReflectionID" json:"recommendations,omitempty"`
}

func (r *Reflection) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

// Answer represents the answers table
type Answer struct {
	ID           uuid.UUID      `gorm:"type:char(36);primaryKey" json:"id"`
	ReflectionID uuid.UUID      `gorm:"type:char(36);not null" json:"reflection_id"`
	QuestionID   uuid.UUID      `gorm:"type:char(36);not null" json:"question_id"`
	OptionID     *uuid.UUID     `gorm:"type:char(36)" json:"option_id"`
	ScaleValue   *int           `gorm:"type:int" json:"scale_value"`
	TextValue    string         `gorm:"type:text" json:"text_value"`
	Score        float64        `gorm:"type:decimal(10,2);default:0" json:"score"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Reflection Reflection `gorm:"foreignKey:ReflectionID" json:"reflection,omitempty"`
	Question   Question   `gorm:"foreignKey:QuestionID" json:"question,omitempty"`
	Option     *Option    `gorm:"foreignKey:OptionID" json:"option,omitempty"`
}

func (a *Answer) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

// CategoryScore represents the category_scores table (Skor per dimensi)
type CategoryScore struct {
	ID           uuid.UUID      `gorm:"type:char(36);primaryKey" json:"id"`
	ReflectionID uuid.UUID      `gorm:"type:char(36);not null" json:"reflection_id"`
	CategoryID   uuid.UUID      `gorm:"type:char(36);not null" json:"category_id"`
	Score        float64        `gorm:"type:decimal(10,2);default:0" json:"score"`
	MaxScore     float64        `gorm:"type:decimal(10,2);default:0" json:"max_score"`
	Percentage   float64        `gorm:"type:decimal(5,2);default:0" json:"percentage"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Reflection Reflection `gorm:"foreignKey:ReflectionID" json:"reflection,omitempty"`
	Category   Category   `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
}

func (cs *CategoryScore) BeforeCreate(tx *gorm.DB) error {
	if cs.ID == uuid.Nil {
		cs.ID = uuid.New()
	}
	return nil
}

// Recommendation represents the recommendations table
type Recommendation struct {
	ID           uuid.UUID      `gorm:"type:char(36);primaryKey" json:"id"`
	ReflectionID uuid.UUID      `gorm:"type:char(36);not null" json:"reflection_id"`
	CategoryID   uuid.UUID      `gorm:"type:char(36);not null" json:"category_id"`
	Title        string         `gorm:"type:varchar(255);not null" json:"title"`
	Description  string         `gorm:"type:text" json:"description"`
	Priority     string         `gorm:"type:varchar(20);default:'medium'" json:"priority"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Reflection Reflection `gorm:"foreignKey:ReflectionID" json:"reflection,omitempty"`
	Category   Category   `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
}

func (r *Recommendation) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

// ScoreThreshold represents the score_thresholds table
type ScoreThreshold struct {
	ID          uuid.UUID      `gorm:"type:char(36);primaryKey" json:"id"`
	CategoryID  *uuid.UUID     `gorm:"type:char(36)" json:"category_id"`
	MinScore    float64        `gorm:"type:decimal(5,2);not null" json:"min_score"`
	MaxScore    float64        `gorm:"type:decimal(5,2);not null" json:"max_score"`
	Label       string         `gorm:"type:varchar(100);not null" json:"label"`
	Description string         `gorm:"type:text" json:"description"`
	Color       string         `gorm:"type:varchar(50)" json:"color"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Category *Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
}

func (st *ScoreThreshold) BeforeCreate(tx *gorm.DB) error {
	if st.ID == uuid.Nil {
		st.ID = uuid.New()
	}
	return nil
}

// TableName overrides
func (User) TableName() string {
	return "users"
}

func (Category) TableName() string {
	return "categories"
}

func (Question) TableName() string {
	return "questions"
}

func (Option) TableName() string {
	return "options"
}

func (Reflection) TableName() string {
	return "reflections"
}

func (Answer) TableName() string {
	return "answers"
}

func (CategoryScore) TableName() string {
	return "category_scores"
}

func (Recommendation) TableName() string {
	return "recommendations"
}

func (ScoreThreshold) TableName() string {
	return "score_thresholds"
}