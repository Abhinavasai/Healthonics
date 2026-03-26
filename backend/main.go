package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/healthonyx/backend/config"
	"github.com/healthonyx/backend/db"
	"github.com/healthonyx/backend/handlers"
	"github.com/healthonyx/backend/middleware"
)

func main() {
	cfg := config.Load()
	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}

	if err := db.Connect(cfg.DatabaseURL); err != nil {
		log.Fatalf("database connection: %v", err)
	}
	defer db.Close()

	if err := db.Migrate(context.Background()); err != nil {
		log.Fatalf("migration: %v", err)
	}

	auth := handlers.NewAuthHandler(cfg.JWTSecret)
	appointments := handlers.NewAppointmentHandler()
	geo := handlers.NewGeoBookingHandler()
	r := gin.Default()

	// CORS: allow Angular dev server and any configured origins
	var origins []string
	if cfg.CORSOrigins != "" {
		for _, o := range strings.Split(cfg.CORSOrigins, ",") {
			if t := strings.TrimSpace(o); t != "" {
				origins = append(origins, t)
			}
		}
	}
	if len(origins) == 0 {
		origins = []string{"http://localhost:4200"}
	}
	r.Use(middleware.CORS(origins))

	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })

	api := r.Group("/api")
	{
		api.POST("/register", auth.Register)
		api.POST("/login", auth.Login)

		// Protected: requires valid JWT
		api.GET("/me", auth.RequireAuth(), auth.Me)

		// Role-protected: demonstrates 403 when role doesn't match
		api.GET("/admin", auth.RequireAuth(), auth.RequireRole("admin"), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "Admin only"})
		})
		api.GET("/doctor", auth.RequireAuth(), auth.RequireRole("doctor"), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "Doctor only"})
		})
		api.GET("/patient", auth.RequireAuth(), auth.RequireRole("patient"), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "Patient only"})
		})

		api.GET("/hospitals/near", auth.RequireAuth(), auth.RequireRole("patient"), geo.ListHospitalsNear)
		api.GET("/doctors/search", auth.RequireAuth(), auth.RequireRole("patient"), geo.SearchDoctors)
		api.GET("/doctors/:id/slots", auth.RequireAuth(), auth.RequireRole("patient"), geo.ListOpenSlotsForDoctor)
		api.POST("/doctor/slots", auth.RequireAuth(), auth.RequireRole("doctor"), geo.CreateSlot)
		api.DELETE("/doctor/slots/:id", auth.RequireAuth(), auth.RequireRole("doctor"), geo.DeleteOpenSlot)
		api.POST("/appointments/book-slot", auth.RequireAuth(), auth.RequireRole("patient"), geo.BookSlot)

		api.POST("/appointments", auth.RequireAuth(), auth.RequireRole("patient"), appointments.Create)
		api.GET("/appointments/patient", auth.RequireAuth(), auth.RequireRole("patient"), appointments.ListPatient)
		api.GET("/appointments/doctor", auth.RequireAuth(), auth.RequireRole("doctor"), appointments.ListDoctor)
		api.GET("/appointments/:id/activity", auth.RequireAuth(), appointments.ListActivity)
		api.GET("/appointments/:id", auth.RequireAuth(), appointments.GetByID)
		api.GET("/doctors", auth.RequireAuth(), auth.RequireRole("patient"), appointments.ListAvailableDoctors)
		api.PATCH("/appointments/:id/status", auth.RequireAuth(), auth.RequireRole("doctor"), appointments.UpdateStatus)
	}

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
