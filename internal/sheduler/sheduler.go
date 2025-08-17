package sheduler

import (
	"context"
	"fmt"
	"readmeow/internal/config"
	"readmeow/internal/domain/repositories"
	"readmeow/pkg/logger"
	"time"

	"github.com/robfig/cron/v3"
)

type Sheduler struct {
	Cron             *cron.Cron
	WidgetRepo       repositories.WidgetRepo
	TemplateRepo     repositories.TemplateRepo
	VerificationRepo repositories.VerificationRepo
	ShedulerConfig   config.ShedulerConfig
	SearchConfig     config.SearchConfig
	Logger           *logger.Logger
}

func NewSheduler(wr repositories.WidgetRepo, tr repositories.TemplateRepo, vr repositories.VerificationRepo, shcfg config.ShedulerConfig, scfg config.SearchConfig, l *logger.Logger) *Sheduler {
	cr := cron.New(cron.WithChain(
		cron.SkipIfStillRunning(cron.DefaultLogger),
	))
	return &Sheduler{
		Cron:             cr,
		WidgetRepo:       wr,
		TemplateRepo:     tr,
		VerificationRepo: vr,
		ShedulerConfig:   shcfg,
		SearchConfig:     scfg,
		Logger:           l,
	}
}

func (s *Sheduler) Start() {
	if _, err := s.Cron.AddFunc(fmt.Sprintf("@every %dm", s.ShedulerConfig.CleanCodesTime), func() {
		op := "sheduler.CleanExpiredVerifyCodes"
		log := s.Logger.AddOp(op)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(s.ShedulerConfig.CleanCodesTimeout))
		defer cancel()
		log.Log.Info("cleaning expired verify codes")
		if err := s.VerificationRepo.DeleteExpired(ctx); err != nil {
			log.Log.Error("failed to delete expired verify codes", logger.Err(err))
		} else {
			log.Log.Info("expired verify codes cleaned successfully")
		}
	}); err != nil {
		panic(fmt.Errorf("failed to start CleanExpiredVerifyCodes sheduler: %w", err))
	}
	if _, err := s.Cron.AddFunc(fmt.Sprintf("@every %dm", s.ShedulerConfig.WidgetBulkTime), func() {
		op := "sheduler.BulkWidgetsData"
		log := s.Logger.AddOp(op)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(s.ShedulerConfig.WidgetBulkTimeout))
		defer cancel()
		log.Log.Info("bulking widgets data")
		if err := s.WidgetRepo.MustBulk(ctx, s.SearchConfig); err != nil {
			log.Log.Error("failed to bulk widgets", logger.Err(err))
		} else {
			log.Log.Info("widgets data bulked successfully")
		}
	}); err != nil {
		panic(fmt.Errorf("failed to start BulkWidgetsData sheduler: %w", err))
	}
	if _, err := s.Cron.AddFunc(fmt.Sprintf("@every %dm", s.ShedulerConfig.TemplateBulkTime), func() {
		op := "sheduler.BulkTemplatesData"
		log := s.Logger.AddOp(op)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(s.ShedulerConfig.TemplateBulkTimeout))
		defer cancel()
		log.Log.Info("bulking templates data")
		if err := s.TemplateRepo.MustBulk(ctx, s.SearchConfig); err != nil {
			log.Log.Error("failed to bulk templates", logger.Err(err))
		} else {
			log.Log.Info("templates data bulked successfully")
		}
	}); err != nil {
		panic(fmt.Errorf("failed to start BulkWidgetsTemplates"))
	}
	s.Cron.Start()
}

func (s *Sheduler) Stop() {
	ctx := s.Cron.Stop()
	<-ctx.Done()
}
