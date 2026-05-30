package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ramisoul84/resreview-server/internal/domain"
	"github.com/ramisoul84/resreview-server/internal/repository"
	"github.com/ramisoul84/resreview-server/pkg/logger"
)

type WsBroadcaster interface {
	Broadcast(data []byte)
}

type AnnotationService interface {
	ListByVersion(ctx context.Context, versionID string) ([]domain.AnnotationResponse, error)
	CreateAnnotation(ctx context.Context, versionID, userID string, req domain.CreateAnnotationRequest) (*domain.AnnotationResponse, error)
	UpdateAnnotation(ctx context.Context, id string, req domain.UpdateAnnotationRequest) error
	DeleteAnnotation(ctx context.Context, id string) error
}

type annotationService struct {
	repo         repository.AnnotationRepository
	versionRepo  repository.VersionRepository
	productRepo  repository.ProductRepository
	ws           WsBroadcaster
	log          logger.Logger
}

func NewAnnotationService(
	repo repository.AnnotationRepository,
	versionRepo repository.VersionRepository,
	productRepo repository.ProductRepository,
	ws WsBroadcaster,
) AnnotationService {
	return &annotationService{
		repo:        repo,
		versionRepo: versionRepo,
		productRepo: productRepo,
		ws:          ws,
		log:         logger.Get(),
	}
}

func (s *annotationService) toResponse(a domain.Annotation) domain.AnnotationResponse {
	return domain.AnnotationResponse{
		ID:          a.ID,
		VersionID:   a.VersionID,
		Type:        a.Type,
		Data:        a.Data,
		UserID:      a.UserID,
		SessionID:   a.SessionID,
		Color:       a.Color,
		StrokeW:     a.StrokeW,
		StrokeStyle: a.StrokeStyle,
		X:           a.X,
		Y:           a.Y,
		Title:       a.Title,
		Text:        a.Text,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
	}
}

func (s *annotationService) ListByVersion(ctx context.Context, versionID string) ([]domain.AnnotationResponse, error) {
	log := s.log.WithFields(map[string]any{
		"layer":      "annotation_service",
		"method":     "ListByVersion",
		"request_id": ctx.Value("request_id"),
	})

	annotations, err := s.repo.ListByVersionID(ctx, versionID)
	if err != nil {
		log.Error().Err(err).Msg("failed to list annotations")
		return nil, err
	}

	resp := make([]domain.AnnotationResponse, len(annotations))
	for i, a := range annotations {
		resp[i] = s.toResponse(a)
	}
	return resp, nil
}

func (s *annotationService) CreateAnnotation(ctx context.Context, versionID, userID string, req domain.CreateAnnotationRequest) (*domain.AnnotationResponse, error) {
	log := s.log.WithFields(map[string]any{
		"layer":      "annotation_service",
		"method":     "CreateAnnotation",
		"request_id": ctx.Value("request_id"),
	})

	now := time.Now()
	ann := &domain.Annotation{
		ID:          uuid.New().String(),
		VersionID:   versionID,
		Type:        req.Type,
		Data:        req.Data,
		UserID:      userID,
		SessionID:   req.SessionID,
		Color:       req.Color,
		StrokeW:     req.StrokeW,
		StrokeStyle: req.StrokeStyle,
		X:           req.X,
		Y:           req.Y,
		Title:       req.Title,
		Text:        req.Text,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.Create(ctx, ann); err != nil {
		log.Error().Err(err).Msg("failed to create annotation")
		return nil, err
	}

	resp := s.toResponse(*ann)

	if s.ws != nil {
		payload, _ := json.Marshal(resp)
		msg := fmt.Appendf(nil, `{"type":"patch","op":{"op":"create_annotation","annotation":%s}}`, payload)
		s.ws.Broadcast(msg)
	}

	log.Debug().Str("annotation_id", ann.ID).Msg("annotation created")
	return &resp, nil
}

func (s *annotationService) UpdateAnnotation(ctx context.Context, id string, req domain.UpdateAnnotationRequest) error {
	log := s.log.WithFields(map[string]any{
		"layer":      "annotation_service",
		"method":     "UpdateAnnotation",
		"request_id": ctx.Value("request_id"),
	})

	ann := &domain.Annotation{
		ID:        id,
		Data:      req.Data,
		X:         req.X,
		Y:         req.Y,
		Title:     req.Title,
		Text:      req.Text,
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Update(ctx, ann); err != nil {
		log.Error().Err(err).Msg("failed to update annotation")
		return err
	}

	if s.ws != nil {
		update := map[string]any{"id": id, "data": req.Data, "x": req.X, "y": req.Y, "title": req.Title, "text": req.Text}
		payload, _ := json.Marshal(update)
		msg := fmt.Appendf(nil, `{"type":"patch","op":{"op":"update_annotation","annotation":%s}}`, payload)
		s.ws.Broadcast(msg)
	}

	log.Debug().Str("annotation_id", id).Msg("annotation updated")
	return nil
}

func (s *annotationService) DeleteAnnotation(ctx context.Context, id string) error {
	log := s.log.WithFields(map[string]any{
		"layer":      "annotation_service",
		"method":     "DeleteAnnotation",
		"request_id": ctx.Value("request_id"),
	})

	if err := s.repo.Delete(ctx, id); err != nil {
		log.Error().Err(err).Msg("failed to delete annotation")
		return err
	}

	if s.ws != nil {
		msg := fmt.Appendf(nil, `{"type":"patch","op":{"op":"delete_annotation","annotationId":"%s"}}`, id)
		s.ws.Broadcast(msg)
	}

	log.Debug().Str("annotation_id", id).Msg("annotation deleted")
	return nil
}
