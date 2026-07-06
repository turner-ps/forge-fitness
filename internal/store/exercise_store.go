// Package store
package store

import "time"

type Exercise struct {
	ID                   int       `json:"id"`
	Name                 string    `json:"name"`
	Level                string    `json:"level"`
	Force                string    `json:"force"`
	Mechanic             string    `json:"mechanic"`
	Equipment            string    `json:"equipement"`
	PrimaryMuscleGroup   []string  `json:"primary_muscle_group"`
	SecondaryMuscleGroup []string  `json:"secondary_muscle_group"`
	Instructions         []string  `json:"Instructions"`
	Category             string    `json:"category"`
	Images               []string  `json:"images"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}
