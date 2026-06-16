package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/healthonyx/backend/config"
	"github.com/healthonyx/backend/db"
	"github.com/healthonyx/backend/handlers"
	"github.com/healthonyx/backend/middleware"
	"github.com/healthonyx/backend/providers"
	"github.com/healthonyx/backend/workers"
)

func main() {
	// Structured JSON logging — all log.Printf calls in workers/handlers will use this.
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))
	// Redirect stdlib log output through slog so third-party libs also produce JSON.
	log.SetFlags(0)

	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		slog.Error("startup configuration invalid", "error", err)
		os.Exit(1)
	}
	handlers.ConfigureAIRuntime(cfg.AIEnabled, cfg.OllamaHost, cfg.OllamaModel, cfg.OllamaTimeoutMS)

	if err := db.Connect(cfg.DatabaseURL); err != nil {
		slog.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Migrate(context.Background()); err != nil {
		slog.Error("database migration failed", "error", err)
		os.Exit(1)
	}

	// Root context cancelled on SIGTERM/SIGINT — shared by handlers, workers, and hub.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	auth := handlers.NewAuthHandler(cfg.JWTSecret)
	bootstrap := handlers.NewBootstrapHandler()
	dashboard := handlers.NewDashboardHandler()
	appointments := handlers.NewAppointmentHandler()
	apptComments := handlers.NewAppointmentCommentsHandler()
	documents := handlers.NewDocumentsHandler()
	doctorDocuments := handlers.NewDoctorDocumentsHandler()
	msgHub := handlers.NewMessagingHub()
	go msgHub.Run(ctx)
	messaging := handlers.NewMessagingHandler(msgHub)
	prescriptions := handlers.NewPrescriptionsHandler()
	notifications := handlers.NewNotificationsHandler()
	providerCallbacks := handlers.NewProviderCallbacksHandler(cfg.SendGridWebhookSecret, cfg.TwilioWebhookSecret)
	criticalEscalations := handlers.NewCriticalEscalationsHandler()
	fhirBoundary := handlers.NewFHIRBoundaryHandler()
	hl7Ingestion := handlers.NewHL7LabIngestionHandler()
	geo := handlers.NewGeoBookingHandler()
	geocode := handlers.NewGeocodeHandler(cfg.NominatimBaseURL, cfg.GeocodeUserAgent, cfg.GoogleMapsAPIKey)
	assistant := handlers.NewAssistantChatHandler(
		cfg.AzureOpenAIEndpoint,
		cfg.AzureOpenAIAPIKey,
		cfg.AzureOpenAIAPIVersion,
		cfg.AzureOpenAIModel,
	)
	aiHealth := handlers.NewAIHealthHandler(
		cfg.AzureOpenAIEndpoint,
		cfg.AzureOpenAIAPIKey,
		cfg.AzureOpenAIAPIVersion,
		cfg.AzureOpenAIModel,
	)
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
	r.Use(middleware.RequestID())
	r.Use(middleware.RequestTelemetry())

	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	r.HEAD("/health", func(c *gin.Context) { c.Status(http.StatusOK) })

	api := r.Group("/api")
	{
		// Rate limit: 10 requests per minute per IP on auth endpoints to block brute-force.
		authLimiter := middleware.NewRateLimiter(ctx, 10, time.Minute)
		api.POST("/register", authLimiter.Middleware(), auth.Register)
		api.POST("/login", authLimiter.Middleware(), auth.Login)
		api.POST("/auth/refresh", authLimiter.Middleware(), auth.Refresh)
		api.POST("/auth/logout", auth.RequireAuth(), auth.Logout)

		// Protected: requires valid JWT
		api.GET("/me", auth.RequireAuth(), auth.Me)
		api.GET("/notifications", auth.RequireAuth(), notifications.ListMine)
		api.POST("/provider-callbacks/sendgrid", providerCallbacks.SendGridWebhook)
		api.POST("/provider-callbacks/twilio", providerCallbacks.TwilioWebhook)
		api.GET("/notifications/preferences", auth.RequireAuth(), notifications.ListPreferences)
		api.PUT("/notifications/preferences", auth.RequireAuth(), notifications.UpsertPreferences)
		api.GET("/bootstrap", auth.RequireAuth(), bootstrap.Get)
		api.GET("/patient/dashboard/summary", auth.RequireAuth(), auth.RequireRole("patient"), dashboard.PatientSummary)
		api.GET("/doctor/dashboard/summary", auth.RequireAuth(), auth.RequireRole("doctor"), dashboard.DoctorSummary)
		api.GET("/doctor/critical-escalations", auth.RequireAuth(), auth.RequireRole("doctor"), criticalEscalations.ListMine)
		api.POST("/doctor/critical-escalations/:id/ack", auth.RequireAuth(), auth.RequireRole("doctor"), criticalEscalations.Ack)
		api.GET("/doctor/documents", auth.RequireAuth(), auth.RequireRole("doctor"), doctorDocuments.List)
		api.GET("/doctor/documents/:id", auth.RequireAuth(), auth.RequireRole("doctor"), doctorDocuments.Get)
		api.GET("/doctor/documents/:id/download", auth.RequireAuth(), auth.RequireRole("doctor"), doctorDocuments.Download)
		api.POST("/doctor/documents/:id/summarize", auth.RequireAuth(), auth.RequireRole("doctor"), doctorDocuments.Summarize)

		// Role-protected: demonstrates 403 when role doesn't match
		adminAudit := handlers.NewAdminAuditHandler()
		knowledge := handlers.NewKnowledgeAdminHandler()
		contextComments := handlers.NewContextualCommentsHandler()
		userLifecycle := handlers.NewAdminUserLifecycleHandler()
		aiRuntime := handlers.NewAdminAIRuntimeHandler()
		api.GET("/admin", auth.RequireAuth(), auth.RequireRole("admin"), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "Admin only"})
		})
		api.GET("/admin/audit-log", auth.RequireAuth(), auth.RequireRole("admin"), adminAudit.List)
		api.GET("/admin/user-lifecycle/users", auth.RequireAuth(), auth.RequireRole("admin"), userLifecycle.ListUsers)
		api.GET("/admin/user-lifecycle/settings-kpis", auth.RequireAuth(), auth.RequireRole("admin"), userLifecycle.GetSettingsKpis)
		api.PUT("/admin/user-lifecycle/settings", auth.RequireAuth(), auth.RequireRole("admin"), userLifecycle.UpdateSettings)
		api.POST("/admin/user-lifecycle/users", auth.RequireAuth(), auth.RequireRole("admin"), userLifecycle.CreateUser)
		api.PUT("/admin/user-lifecycle/users/:id", auth.RequireAuth(), auth.RequireRole("admin"), userLifecycle.UpdateUser)
		api.PATCH("/admin/user-lifecycle/users/:id/deactivate", auth.RequireAuth(), auth.RequireRole("admin"), userLifecycle.DeactivateUser)
		api.POST("/admin/user-lifecycle/users/:id/reset-password", auth.RequireAuth(), auth.RequireRole("admin"), userLifecycle.ResetPassword)
		api.GET("/admin/ai/observability", auth.RequireAuth(), auth.RequireRole("admin"), aiRuntime.GetObservability)
		api.PUT("/admin/ai/settings", auth.RequireAuth(), auth.RequireRole("admin"), aiRuntime.UpdateSettings)
		api.POST("/admin/ai/eval", auth.RequireAuth(), auth.RequireRole("admin"), aiRuntime.RunEval)
		api.GET("/admin/fhir/export", auth.RequireAuth(), auth.RequireRole("admin"), fhirBoundary.ExportBundle)
		api.POST("/admin/fhir/import", auth.RequireAuth(), auth.RequireRole("admin"), fhirBoundary.ImportBundle)
		api.POST("/integrations/hl7/lab-results", auth.RequireAuth(), auth.RequireRole("admin"), hl7Ingestion.IngestLabResult)
		api.GET("/admin/notifications", auth.RequireAuth(), auth.RequireRole("admin"), notifications.AdminList)
		api.GET("/admin/notifications/summary", auth.RequireAuth(), auth.RequireRole("admin"), notifications.AdminSummary)
		api.GET("/admin/notifications/consent-history", auth.RequireAuth(), auth.RequireRole("admin"), notifications.AdminConsentHistory)
		api.POST("/admin/notifications/:id/retry", auth.RequireAuth(), auth.RequireRole("admin"), notifications.RetryFailed)
		api.GET("/admin/knowledge-docs", auth.RequireAuth(), auth.RequireRole("admin"), knowledge.List)
		api.POST("/admin/knowledge-docs", auth.RequireAuth(), auth.RequireRole("admin"), knowledge.Create)
		api.GET("/admin/knowledge-docs/similarity-scan", auth.RequireAuth(), auth.RequireRole("admin"), knowledge.SimilarityScan)
		api.POST("/admin/knowledge-docs/:id/embeddings", auth.RequireAuth(), auth.RequireRole("admin"), knowledge.UpsertEmbeddings)
		api.GET("/admin/knowledge-docs/:id/versions", auth.RequireAuth(), auth.RequireRole("admin"), knowledge.ListVersions)
		api.POST("/admin/knowledge-docs/:id/review", auth.RequireAuth(), auth.RequireRole("admin"), knowledge.Review)
		api.PATCH("/admin/knowledge-docs/:id", auth.RequireAuth(), auth.RequireRole("admin"), knowledge.Patch)
		api.GET("/admin/knowledge-docs/:id", auth.RequireAuth(), auth.RequireRole("admin"), knowledge.Get)
		api.GET("/doctor", auth.RequireAuth(), auth.RequireRole("doctor"), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "Doctor only"})
		})
		api.GET("/patient", auth.RequireAuth(), auth.RequireRole("patient"), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "Patient only"})
		})

		api.GET("/geocode", auth.RequireAuth(), auth.RequireRole("patient"), geocode.GeocodeSearch)
		api.POST("/find-care/places/nearby", auth.RequireAuth(), auth.RequireRole("patient"), geocode.NearbyHospitalsGoogle)
		api.GET("/hospitals/near", auth.RequireAuth(), auth.RequireRole("patient"), geo.ListHospitalsNear)
		api.GET("/doctors/search", auth.RequireAuth(), auth.RequireRole("patient"), geo.SearchDoctors)
		api.GET("/doctors/:id/slots", auth.RequireAuth(), auth.RequireRole("patient", "doctor"), geo.ListOpenSlotsForDoctor)
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
		api.GET("/comments/:contextType/:contextId", auth.RequireAuth(), contextComments.List)
		api.POST("/comments/:contextType/:contextId", auth.RequireAuth(), contextComments.Create)
		api.GET("/appointments/:id", auth.RequireAuth(), appointments.GetByID)
		api.GET("/doctors", auth.RequireAuth(), auth.RequireRole("patient"), appointments.ListAvailableDoctors)
		api.PATCH("/appointments/:id/status", auth.RequireAuth(), auth.RequireRole("doctor"), appointments.UpdateStatus)

		api.GET("/documents", auth.RequireAuth(), auth.RequireRole("patient"), documents.List)
		api.POST("/documents", auth.RequireAuth(), auth.RequireRole("patient"), documents.Upload)
		api.GET("/documents/:id/download", auth.RequireAuth(), auth.RequireRole("patient"), documents.Download)

		api.GET("/patients/:patientId/files", auth.RequireAuth(), patientFiles.List)
		api.POST("/patients/:patientId/files", auth.RequireAuth(), patientFiles.Upload)
		api.GET("/files/:id", auth.RequireAuth(), patientFiles.Download)

		api.GET("/patients/:patientId/prescriptions", auth.RequireAuth(), prescriptions.ListByPatient)
		api.POST("/patients/:patientId/prescriptions", auth.RequireAuth(), auth.RequireRole("doctor"), prescriptions.Create)
		api.GET("/prescriptions/:id/pdf", auth.RequireAuth(), auth.RequireRole("patient", "doctor", "admin"), prescriptions.DownloadPDF)
		api.PUT("/prescriptions/:id", auth.RequireAuth(), auth.RequireRole("doctor", "admin"), prescriptions.Update)
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
		api.GET("/messages/ws", messaging.ServeWebSocket(auth))
		api.POST("/assistant/chat", auth.RequireAuth(), assistant.Chat)
		api.POST("/assistant/symptom-check", auth.RequireAuth(), auth.RequireRole("patient"), aiHealth.SymptomCheck)
		api.POST("/assistant/drug-interactions", auth.RequireAuth(), auth.RequireRole("doctor", "admin"), aiHealth.DrugInteractions)
		api.GET("/patients/:patientId/ai-summary", auth.RequireAuth(), auth.RequireRole("doctor", "admin"), aiHealth.PatientSummary)
		api.POST("/appointments/:id/note-assist", auth.RequireAuth(), auth.RequireRole("doctor", "admin"), aiHealth.NoteAssist)
		api.GET("/assistant/symptom-trends", auth.RequireAuth(), auth.RequireRole("patient"), aiHealth.SymptomTrends)
		slotRec := handlers.NewSlotRecommenderHandler()
		api.GET("/appointments/recommend", auth.RequireAuth(), auth.RequireRole("patient"), slotRec.Recommend)
		preVisit := handlers.NewPreVisitHandler(aiHealth)
		api.GET("/appointments/:id/questionnaire", auth.RequireAuth(), preVisit.GetQuestionnaire)
		api.POST("/appointments/:id/questionnaire", auth.RequireAuth(), auth.RequireRole("patient"), preVisit.SubmitAnswers)
		adherence := handlers.NewAdherenceHandler()
		api.GET("/patients/:patientId/adherence-score", auth.RequireAuth(), adherence.Score)
		heatmap := handlers.NewWorkloadHeatmapHandler()
		api.GET("/doctor/workload-heatmap", auth.RequireAuth(), auth.RequireRole("doctor", "admin"), heatmap.Get)
		fhirExport := handlers.NewFHIRExportHandler()
		api.GET("/patient/fhir-export", auth.RequireAuth(), auth.RequireRole("patient"), fhirExport.Export)
		so := handlers.NewSecondOpinionHandler()
		api.GET("/second-opinions", auth.RequireAuth(), auth.RequireRole("doctor", "admin"), so.List)
		api.POST("/second-opinions", auth.RequireAuth(), auth.RequireRole("doctor", "admin"), so.Create)
		api.GET("/second-opinions/:id/replies", auth.RequireAuth(), auth.RequireRole("doctor", "admin"), so.GetReplies)
		api.POST("/second-opinions/:id/replies", auth.RequireAuth(), auth.RequireRole("doctor", "admin"), so.PostReply)
		api.PATCH("/second-opinions/:id/resolve", auth.RequireAuth(), auth.RequireRole("doctor", "admin"), so.Resolve)

		// Feature 1: Teleconsult video link
		api.PATCH("/appointments/:id/video-link", auth.RequireAuth(), auth.RequireRole("doctor"), appointments.SetVideoLink)

		// Feature 2: Prescription refill requests
		refill := handlers.NewRefillRequestHandler()
		api.POST("/prescriptions/:id/refill-request", auth.RequireAuth(), auth.RequireRole("patient"), refill.PatientCreate)
		api.GET("/patient/refill-requests", auth.RequireAuth(), auth.RequireRole("patient"), refill.PatientList)
		api.GET("/doctor/refill-requests", auth.RequireAuth(), auth.RequireRole("doctor"), refill.DoctorList)
		api.PATCH("/refill-requests/:id/review", auth.RequireAuth(), auth.RequireRole("doctor"), refill.DoctorReview)

		// Feature 3: Lab results (AI interpretation on existing patient_documents)
		labResults := handlers.NewLabResultsHandler(aiHealth)
		api.GET("/patient/lab-results", auth.RequireAuth(), auth.RequireRole("patient"), labResults.PatientList)
		api.POST("/patient/lab-results/:id/interpret", auth.RequireAuth(), auth.RequireRole("patient"), labResults.Interpret)

		// Feature 4: Doctor patient panel
		doctorPatients := handlers.NewDoctorPatientsHandler()
		api.GET("/doctor/patients", auth.RequireAuth(), auth.RequireRole("doctor"), doctorPatients.List)

		// Feature 6: Patient health summary export
		healthSummary := handlers.NewHealthSummaryPDFHandler()
		api.GET("/patient/health-summary", auth.RequireAuth(), auth.RequireRole("patient"), healthSummary.Export)

		// Feature 7: Admin SLO dashboard
		slo := handlers.NewAdminSLOHandler()
		api.GET("/admin/slo", auth.RequireAuth(), auth.RequireRole("admin"), slo.Dashboard)

		// Feature 8: Bulk messaging
		api.POST("/doctor/broadcast", auth.RequireAuth(), auth.RequireRole("doctor"), messaging.BulkBroadcast)

		// Feature 9: Waiting room
		waitingRoom := handlers.NewWaitingRoomHandler()
		api.GET("/doctor/waiting-room", auth.RequireAuth(), auth.RequireRole("doctor"), waitingRoom.Queue)
		api.POST("/appointments/:id/check-in", auth.RequireAuth(), auth.RequireRole("patient"), waitingRoom.CheckIn)

		// Feature 10: Appointment ratings
		ratings := handlers.NewAppointmentRatingHandler()
		api.POST("/appointments/:id/rating", auth.RequireAuth(), auth.RequireRole("patient"), ratings.Submit)
		api.GET("/appointments/:id/rating", auth.RequireAuth(), ratings.GetForAppointment)
		api.GET("/doctor/rating-summary", auth.RequireAuth(), auth.RequireRole("doctor"), ratings.DoctorRatingSummary)
	}

	addr := fmt.Sprintf(":%s", cfg.Port)
	go workers.NewDocumentSummaryWorker(2 * time.Second).Run(ctx)
	workers.StartAnomalyWorker(ctx)
	go workers.NewAppointmentReminderWorker(5 * time.Minute).Run(ctx)
	notificationProviders := workers.NewNotificationProviderSet(
		providers.NewSendGridAdapter(cfg.SendGridAPIKey, cfg.SendGridFrom),
		providers.NewTwilioAdapter(cfg.TwilioAccountSID, cfg.TwilioAuthToken, cfg.TwilioFromNumber),
	)
	go workers.NewNotificationWorkerWithConfig(workers.NotificationWorkerConfig{
		Interval:   time.Duration(cfg.NotifyIntervalMS) * time.Millisecond,
		MaxRetries: cfg.NotifyMaxRetries,
		Providers:  notificationProviders,
	}).Run(ctx)

	srv := &http.Server{Addr: addr, Handler: r}
	go func() {
		<-ctx.Done()
		slog.Info("shutting down gracefully")
		shutCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutCtx)
	}()

	slog.Info("server starting", "addr", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
