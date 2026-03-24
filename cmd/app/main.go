package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"room-booking/internal/auth"
	"room-booking/internal/db"
	"room-booking/internal/middleware"
	"room-booking/internal/repository"
)

func main() {
	database := db.New()
	roomRepo := repository.NewRoomRepository(database)

	database.Exec(`
	CREATE TABLE IF NOT EXISTS schedules (
		id SERIAL PRIMARY KEY,
		room_id INT,
		days TEXT,
		start_time TEXT,
		end_time TEXT
	)
	`)

	database.Exec(`
	CREATE TABLE IF NOT EXISTS bookings (
		id SERIAL PRIMARY KEY,
		user_id TEXT,
		room_id INT,
		start_time TEXT,
		end_time TEXT,
		status TEXT
	)
	`)

	// INFO
	http.HandleFunc("/_info", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	// AUTH
	http.HandleFunc("/dummyLogin", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Role string `json:"role"`
		}
		json.NewDecoder(r.Body).Decode(&req)

		token, _ := auth.GenerateToken(req.Role)

		json.NewEncoder(w).Encode(map[string]string{
			"token": token,
		})
	})

	// ROOMS
	http.HandleFunc("/rooms/create", middleware.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		role := r.Context().Value(middleware.RoleKey).(string)
		if role != "admin" {
			http.Error(w, "forbidden", 403)
			return
		}

		var room repository.Room
		json.NewDecoder(r.Body).Decode(&room)
		roomRepo.Create(room)

		json.NewEncoder(w).Encode(map[string]interface{}{"room": room})
	}))

	http.HandleFunc("/rooms/list", middleware.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		rooms, _ := roomRepo.GetAll()
		json.NewEncoder(w).Encode(map[string]interface{}{"rooms": rooms})
	}))

	// SCHEDULE
	http.HandleFunc("/schedules/create", middleware.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		role := r.Context().Value(middleware.RoleKey).(string)
		if role != "admin" {
			http.Error(w, "forbidden", 403)
			return
		}

		var req struct {
			RoomID    int   `json:"room_id"`
			Days      []int `json:"days"`
			StartTime string `json:"start_time"`
			EndTime   string `json:"end_time"`
		}

		json.NewDecoder(r.Body).Decode(&req)

		var parts []string
		for _, d := range req.Days {
			parts = append(parts, strconv.Itoa(d))
		}

		database.Exec(
			"INSERT INTO schedules (room_id, days, start_time, end_time) VALUES ($1,$2,$3,$4)",
			req.RoomID, strings.Join(parts, ","), req.StartTime, req.EndTime,
		)

		json.NewEncoder(w).Encode(map[string]interface{}{"schedule": req})
	}))

	// SLOTS
	http.HandleFunc("/slots", middleware.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {

		roomID, _ := strconv.Atoi(r.URL.Query().Get("room_id"))
		date, _ := time.Parse("2006-01-02", r.URL.Query().Get("date"))

		row := database.QueryRow("SELECT days, start_time, end_time FROM schedules WHERE room_id=$1 LIMIT 1", roomID)

		var daysStr, startTime, endTime string
		err := row.Scan(&daysStr, &startTime, &endTime)
		if err != nil {
			http.Error(w, "no schedule", 404)
			return
		}

		var days []int
		for _, p := range strings.Split(daysStr, ",") {
			v, _ := strconv.Atoi(p)
			days = append(days, v)
		}

		weekday := int(date.Weekday())

		found := false
		for _, d := range days {
			if d == weekday {
				found = true
				break
			}
		}

		if !found {
			json.NewEncoder(w).Encode(map[string]interface{}{"slots": []interface{}{}})
			return
		}

		layout := "15:04"
		start, _ := time.Parse(layout, startTime)
		end, _ := time.Parse(layout, endTime)

		var slots []map[string]string

		for start.Before(end) {
			next := start.Add(30 * time.Minute)

			slots = append(slots, map[string]string{
				"start": start.Format(layout),
				"end":   next.Format(layout),
			})

			start = next
		}

		json.NewEncoder(w).Encode(map[string]interface{}{"slots": slots})
	}))

	// BOOKING CREATE
	http.HandleFunc("/bookings/create", middleware.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {

		role := r.Context().Value(middleware.RoleKey).(string)
		if role != "user" {
			http.Error(w, "only user", 403)
			return
		}

		userID := r.Context().Value(middleware.UserIDKey).(string)

		var req struct {
			RoomID    int    `json:"room_id"`
			StartTime string `json:"start_time"`
			EndTime   string `json:"end_time"`
		}

		json.NewDecoder(r.Body).Decode(&req)

		layout := "15:04"

		now := time.Now()
		slotTime, _ := time.Parse(layout, req.StartTime)

		current := time.Date(0, 1, 1, now.Hour(), now.Minute(), 0, 0, time.UTC)
		slot := time.Date(0, 1, 1, slotTime.Hour(), slotTime.Minute(), 0, 0, time.UTC)

		if slot.Before(current) {
			http.Error(w, "cannot book past", 400)
			return
		}

		var count int
		database.QueryRow(
			"SELECT COUNT(*) FROM bookings WHERE room_id=$1 AND start_time=$2 AND end_time=$3 AND status='active'",
			req.RoomID, req.StartTime, req.EndTime,
		).Scan(&count)

		if count > 0 {
			http.Error(w, "slot already booked", 400)
			return
		}

		database.Exec(
			"INSERT INTO bookings (user_id, room_id, start_time, end_time, status) VALUES ($1,$2,$3,$4,'active')",
			userID, req.RoomID, req.StartTime, req.EndTime,
		)

		json.NewEncoder(w).Encode(map[string]string{"status": "booked"})
	}))

	// BOOKING CANCEL (идемпотентно)
	http.HandleFunc("/bookings/cancel", middleware.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {

		userID := r.Context().Value(middleware.UserIDKey).(string)

		var req struct {
			RoomID    int    `json:"room_id"`
			StartTime string `json:"start_time"`
			EndTime   string `json:"end_time"`
		}

		json.NewDecoder(r.Body).Decode(&req)

		database.Exec(
			"UPDATE bookings SET status='cancelled' WHERE user_id=$1 AND room_id=$2 AND start_time=$3 AND end_time=$4",
			userID, req.RoomID, req.StartTime, req.EndTime,
		)

		json.NewEncoder(w).Encode(map[string]string{"status": "cancelled"})
	}))

	// MY BOOKINGS (только будущие)
	http.HandleFunc("/bookings/my", middleware.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {

		userID := r.Context().Value(middleware.UserIDKey).(string)

		rows, _ := database.Query(
			"SELECT room_id, start_time, end_time, status FROM bookings WHERE user_id=$1 AND status='active'",
			userID,
		)

		now := time.Now()
		layout := "15:04"

		var result []map[string]interface{}

		for rows.Next() {
			var roomID int
			var start, end, status string

			rows.Scan(&roomID, &start, &end, &status)

			t, _ := time.Parse(layout, start)

			slot := time.Date(0, 1, 1, t.Hour(), t.Minute(), 0, 0, time.UTC)
			current := time.Date(0, 1, 1, now.Hour(), now.Minute(), 0, 0, time.UTC)

			if slot.Before(current) {
				continue
			}

			result = append(result, map[string]interface{}{
				"room_id": roomID,
				"start":   start,
				"end":     end,
				"status":  status,
			})
		}

		json.NewEncoder(w).Encode(map[string]interface{}{"bookings": result})
	}))

	log.Println("server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}