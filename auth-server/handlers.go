package main

import (
	"context"
	"net/http"
	"time"
	"log"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

func signupHandler(c *gin.Context) {
	var req SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	// Check if user already exists
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	var existingUser User
	err := usersColl.FindOne(ctx, bson.M{"email": req.Email}).Decode(&existingUser)
	if err == nil {
		c.JSON(http.StatusConflict, ErrorResponse{
			Error:   "user_exists",
			Message: "User with this email already exists",
		})
		return
	}

	// Hash Password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to process password",
		})
		return
	}

	// Create New User
	user := User{
		Email:     req.Email,
		Password:  string(hashedPassword),
		FirstName: req.FirstName,
		LastName:  req.LastName,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	result, err := usersColl.InsertOne(ctx, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to create user",
		})
		return
	}

	user.ID = result.InsertedID.(primitive.ObjectID)

	// Generate JWT Token
	token, err := generateJWT(user.ID.Hex())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "token_error",
			Message: "Failed to generate authentication token",
		})
		return
	}

	c.JSON(http.StatusCreated, LoginResponse{
		User:  user,
		Token: token,
	})
}

func loginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	// Find user by email
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	var user User
	err := usersColl.FindOne(ctx, bson.M{"email": req.Email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error:   "invalid_credentials",
				Message: "Invalid email or password",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to authenticate user",
		})
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "invalid_credentials",
			Message: "Invalid email or password",
		})
		return
	}

	// Generate JWT token
	token, err := generateJWT(user.ID.Hex())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "token_error",
			Message: "Failed to generate authentication token",
		})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		User:  user,
		Token: token,
	})
}

func getReportsHandler(c *gin.Context) {	
	userID := c.GetString("userID")
	log.Println("[DEBUG] userID from middleware:", userID)

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	var reports []Report
	filter := bson.M{"user_id": userID}

	cursor, err := reportsColl.Find(ctx, filter)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "no_report_found",
				Message: "No report found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to fetch user reports",
		})
		return
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var report Report
		if err := cursor.Decode(&report); err != nil {
			log.Println("[ERROR] Failed to decode report:", err)
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "database_error",
				Message: "Failed to decode report",
			})
			return
		}
		reports = append(reports, report)
	}

	if err := cursor.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to fetch user reports",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Reports fetched successfully",
		Data:    reports,
	})
}

// New handler to get user information
func getUserInfoHandler(c *gin.Context) {
	userID := c.GetString("userID")
	log.Println("[DEBUG] Getting user info for userID:", userID)

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	// Convert userID string to ObjectID
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID format",
		})
		return
	}

	// Find user by ID
	var user User
	err = usersColl.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "user_not_found",
				Message: "User not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to fetch user information",
		})
		return
	}

	// Count number of reports (attacks) for this user
	reportsCount, err := reportsColl.CountDocuments(ctx, bson.M{"user_id": userID})
	if err != nil {
		log.Println("[ERROR] Failed to count reports:", err)
		// Continue even if count fails, just set to 0
		reportsCount = 0
	}

	// Prepare response with user info
	userInfo := map[string]interface{}{
		"user_id":       user.ID.Hex(),
		"email":         user.Email,
		"first_name":    user.FirstName,
		"last_name":     user.LastName,
		"username":      user.FirstName + " " + user.LastName,
		"attacks_count": reportsCount,
		"created_at":    user.CreatedAt,
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "User information fetched successfully",
		Data:    userInfo,
	})
}
