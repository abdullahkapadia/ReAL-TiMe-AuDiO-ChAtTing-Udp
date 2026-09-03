package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"go-udp-server/api"
	"go-udp-server/api/models"
)

type FriendRequest struct {
	RequesterID uint `json:"requesterId"`
	AddresseeID uint `json:"addresseeId"`
}

// SendFriendRequest handles POST /friends/request
func SendFriendRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req FriendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	friendship := models.Friendship{
		RequesterID: req.RequesterID,
		AddresseeID: req.AddresseeID,
		Status:      "pending",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	result := api.DB.Create(&friendship)
	if result.Error != nil {
		http.Error(w, "Failed to send friend request", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Friend request sent!"})
}

// AcceptFriendRequest handles POST /friends/accept
func AcceptFriendRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req FriendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Find the pending request
	var friendship models.Friendship
	result := api.DB.Where("requester_id = ? AND addressee_id = ? AND status = ?", req.RequesterID, req.AddresseeID, "pending").First(&friendship)
	if result.Error != nil {
		http.Error(w, "Friend request not found", http.StatusNotFound)
		return
	}

	// Update status
	friendship.Status = "accepted"
	api.DB.Save(&friendship)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Friend request accepted!"})
}

// GetFriends handles GET /friends?userId=1
func GetFriends(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userId := r.URL.Query().Get("userId")
	if userId == "" {
		http.Error(w, "Missing userId parameter", http.StatusBadRequest)
		return
	}

	var friendships []models.Friendship
	// Find where user is either requester or addressee and status is accepted
	api.DB.Where("(requester_id = ? OR addressee_id = ?) AND status = ?", userId, userId, "accepted").Find(&friendships)

	json.NewEncoder(w).Encode(friendships)
}
