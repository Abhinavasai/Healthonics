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
	bootstrap := handlers.NewBootstrapHandler()
	dashboard := handlers.NewDashboardHandler()
	appointments := handlers.NewAppointmentHandler()
	apptComments := handlers.NewAppointmentCommentsHandler()
	messaging := handlers.NewMessagingHandler()
	prescriptions := handlers.NewPrescriptionsHandler()
	notifications := handlers.NewNotificationsHandler()
	geo := handlers.NewGeoBookingHandler()
	geocode := handlers.NewGeocodeHandler(cfg.NominatimBaseURL, cfg.GeocodeUserAgent)
	patientFiles := handlers.NewPatientFilesHandler(cfg.UploadDir)
	r := gin.Default()
	r.MaxMultipartMemory = 8 << 20 // 8 MiB multipart buffer (handler still enforces 5 MiB file cap)

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
		api.GET("/notifications", auth.RequireAuth(), notifications.ListMine)
		api.GET("/bootstrap", auth.RequireAuth(), bootstrap.Get)
		api.GET("/patient/dashboard/summary", auth.RequireAuth(), auth.RequireRole("patient"), dashboard.PatientSummary)
		api.GET("/doctor/dashboard/summary", auth.RequireAuth(), auth.RequireRole("doctor"), dashboard.DoctorSummary)

		// Role-protected: demonstrates 403 when role doesn't match
		adminAudit := handlers.NewAdminAuditHandler()
		knowledge := handlers.NewKnowledgeAdminHandler()
		api.GET("/admin", auth.RequireAuth(), auth.RequireRole("admin"), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "Admin only"})
		})
		api.GET("/admin/audit-log", auth.RequireAuth(), auth.RequireRole("admin"), adminAudit.List)
		api.GET("/admin/knowledge-docs", auth.RequireAuth(), auth.RequireRole("admin"), knowledge.List)
		api.POST("/admin/knowledge-docs", auth.RequireAuth(), auth.RequireRole("admin"), knowledge.Create)
		api.GET("/admin/knowledge-docs/:id", auth.RequireAuth(), auth.RequireRole("admin"), knowledge.Get)
		api.GET("/doctor", auth.RequireAuth(), auth.RequireRole("doctor"), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "Doctor only"})
		})
		api.GET("/patient", auth.RequireAuth(), auth.RequireRole("patient"), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "Patient only"})
		})

		api.GET("/geocode", auth.RequireAuth(), auth.RequireRole("patient"), geocode.GeocodeSearch)
		api.GET("/hospitals/near", auth.RequireAuth(), auth.RequireRole("patient"), geo.ListHospitalsNear)
		api.GET("/doctors/search", auth.RequireAuth(), auth.RequireRole("patient"), geo.SearchDoctors)
		api.GET("/doctors/:id/slots", auth.RequireAuth(), auth.RequireRole("patient"), geo.ListOpenSlotsForDoctor)
		api.POST("/doctor/slots", auth.RequireAuth(), auth.RequireRole("doctor"), geo.CreateSlot)
		api.DELETE("/doctor/slots/:id", auth.RequireAuth(), auth.RequireRole("doctor"), geo.DeleteOpenSlot)
		api.POST("/appointments/book-slot", auth.RequireAuth(), auth.RequireRole("patient"), geo.BookSlot)

		api.POST("/appointments", auth.RequireAuth(), auth.RequireRole("patient"), appointments.Create)
		api.POST("/patient/appointments/:id/cancel", auth.RequireAuth(), auth.RequireRole("patient"), appointments.PatientCancel)
		api.GET("/appointments/patient", auth.RequireAuth(), auth.RequireRole("patient"), appointments.ListPatient)
		api.GET("/appointments/doctor", auth.RequireAuth(), auth.RequireRole("doctor"), appointments.ListDoctor)
		api.GET("/appointments/:id/activity", auth.RequireAuth(), appointments.ListActivity)
		api.GET("/appointments/:id/comments", auth.RequireAuth(), apptComments.List)
		api.POST("/appointments/:id/comments", auth.RequireAuth(), apptComments.Create)
		api.GET("/appointments/:id", auth.RequireAuth(), appointments.GetByID)
		api.GET("/doctors", auth.RequireAuth(), auth.RequireRole("patient"), appointments.ListAvailableDoctors)
		api.PATCH("/appointments/:id/status", auth.RequireAuth(), auth.RequireRole("doctor"), appointments.UpdateStatus)

		api.GET("/patients/:patientId/files", auth.RequireAuth(), patientFiles.List)
		api.POST("/patients/:patientId/files", auth.RequireAuth(), patientFiles.Upload)
		api.GET("/files/:id", auth.RequireAuth(), patientFiles.Download)

		api.GET("/patients/:patientId/prescriptions", auth.RequireAuth(), prescriptions.ListByPatient)
		api.POST("/patients/:patientId/prescriptions", auth.RequireAuth(), auth.RequireRole("doctor"), prescriptions.Create)
		api.PATCH("/prescriptions/:id/revoke", auth.RequireAuth(), auth.RequireRole("doctor", "admin"), prescriptions.Revoke)

		msg := api.Group("/messages", auth.RequireAuth(), auth.RequireRole("patient", "doctor"))
		{
			msg.GET("/unread", messaging.UnreadTotal)
			msg.GET("/threads", messaging.ListThreads)
			msg.POST("/threads", messaging.CreateThread)
			msg.GET("/threads/:threadId", messaging.ListMessages)
			msg.POST("/threads/:threadId/messages", messaging.SendMessage)
			msg.POST("/threads/:threadId/read", messaging.MarkRead)
		}
	}

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
