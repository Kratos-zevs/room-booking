package repository

import (
	"database/sql"
)

type Room struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Capacity    int    `json:"capacity"`
}

type RoomRepository struct {
	db *sql.DB
}

func NewRoomRepository(db *sql.DB) *RoomRepository {
	return &RoomRepository{db: db}
}

func (r *RoomRepository) Create(room Room) error {
	_, err := r.db.Exec(
		"INSERT INTO rooms (name, description, capacity) VALUES ($1, $2, $3)",
		room.Name, room.Description, room.Capacity,
	)
	return err
}

func (r *RoomRepository) GetAll() ([]Room, error) {
	rows, err := r.db.Query("SELECT id, name, description, capacity FROM rooms")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []Room

	for rows.Next() {
		var r Room
		err := rows.Scan(&r.ID, &r.Name, &r.Description, &r.Capacity)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, r)
	}

	return rooms, nil
}