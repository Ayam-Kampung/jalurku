package model

import (
	"time"
	"log"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Jurusan seperti TJA, TKJ, RPL, PG
type Jurusan struct {
	ID         	int            		`gorm:"primaryKey;autoIncrement" json:"id"`
	Name	   	string         		`gorm:"type:varchar(50);unique;not null" json:"name"` 
	CreatedAt  	time.Time
	UpdatedAt  	time.Time

	// Relasi
	Pertanyaan  []Pertanyaan  		`gorm:"foreignKey:JurusanID"`
	HasilAngket []HasilAngket 		`gorm:"foreignKey:JurusanID"`
}

// Bentuk pertanyaan yang berhubungan dengan jurusan
type Pertanyaan struct {
	ID        	uuid.UUID           `gorm:"type:char(36);primaryKey" json:"id"`
	Text      	string         		`gorm:"type:text;not null" json:"text"`
	Image		string				`json:"image"`
	JurusanID 	int            		`gorm:"not null" json:"jurusan_id"`
	CreatedAt 	time.Time
	UpdatedAt 	time.Time

	Jurusan 	Jurusan 			`gorm:"foreignKey:JurusanID"`
}

// Hasil angket yang berhubungan dengan pengguna
type HasilAngket struct {
	ID        	uuid.UUID      		`gorm:"type:char(36);primaryKey" json:"id"`
	UserID    	uuid.UUID      		`gorm:"type:char(36);not null" json:"user_id"`
	JurusanID 	int      		    `gorm:"not null" json:"jurusan_id"` // Ubah ke int
	CreatedAt 	time.Time
	UpdatedAt 	time.Time
	DeletedAt 	gorm.DeletedAt 		`gorm:"index"`

	User    	User    			`gorm:"foreignKey:UserID"`
	Jurusan 	Jurusan 			`gorm:"foreignKey:JurusanID"`
}

type SubmitRequest struct {
	SessionID       string `json:"session_id"`
	QuestionID   string `json:"question_id"`
	SelectedOption int   `json:"selected_option"`
}

// Tambahkan data Jurusan -> (1:PG, 2:RPL, 3:TKJ, 4:TJA)
func SeedJurusan(db *gorm.DB) {
	// Hapus semua data jurusan
	db.Exec("DELETE FROM jurusan")
	
	jurusanData := []Jurusan{
		{ID: 1, Name: "PG"	},
		{ID: 2, Name: "RPL"	},
		{ID: 3, Name: "TKJ"	},
		{ID: 4, Name: "TJA"	},
	}

	if err := db.Create(&jurusanData).Error; err != nil {
		log.Printf("Error re-seeding jurusan: %v", err)
	} else {
		log.Println("Jurusan re-seeded successfully!")
	}
}

func (Jurusan) TableName() string {
	return "jurusan"
}

func (Pertanyaan) TableName() string {
	return "pertanyaan"
}

func (HasilAngket) TableName() string {
	return "hasil_angket"
}