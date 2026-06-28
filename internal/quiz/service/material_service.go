package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"

	"github.com/pranavbh-9117/IMB/internal/domain"
	"github.com/pranavbh-9117/IMB/internal/quiz/dto"
	"github.com/pranavbh-9117/IMB/internal/quiz/repository"
	"github.com/pranavbh-9117/IMB/pkg/apperror"
	"github.com/pranavbh-9117/IMB/pkg/logger"
	"github.com/pranavbh-9117/IMB/pkg/storage"
)

var (
	ErrUnsupportedFileType  = apperror.BadRequest("unsupported file type")
	ErrFileTooLarge         = apperror.BadRequest("file exceeds maximum limit of 20 MB")
	ErrMaterialLimitReached = apperror.BadRequest("quiz material limit of 10 reached")
	ErrMaterialNotFound     = apperror.NotFound("quiz material not found")
)

type materialService struct {
	quizRepo     repository.QuizRepository
	materialRepo repository.MaterialRepository
	storage      storage.Storage
}

// NewMaterialService initializes a MaterialService.
func NewMaterialService(quizRepo repository.QuizRepository, materialRepo repository.MaterialRepository, st storage.Storage) MaterialService {
	return &materialService{
		quizRepo:     quizRepo,
		materialRepo: materialRepo,
		storage:      st,
	}
}

func (s *materialService) authorizeQuizAccess(ctx context.Context, quiz *domain.Quiz, institutionID uuid.UUID, callerID uuid.UUID, callerRole domain.Role) error {
	if callerRole == domain.RoleFaculty {
		if quiz.CreatedBy != callerID {
			return ErrUnauthorizedQuiz
		}
	} else if callerRole == domain.RoleStudent {
		if quiz.InstitutionID != institutionID || !quiz.IsPublished {
			return ErrUnauthorizedQuiz
		}
	} else {
		return ErrUnauthorizedQuiz
	}
	return nil
}

func (s *materialService) UploadMaterials(ctx context.Context, institutionID uuid.UUID, facultyID uuid.UUID, quizID uuid.UUID, files []*multipart.FileHeader) (*dto.UploadMaterialsResponse, error) {
	// 1. Validate quiz ownership and ensure the material count limit is not exceeded.
	quiz, err := s.quizRepo.GetQuizByID(ctx, quizID)
	if err != nil {
		return nil, err
	}

	if quiz.InstitutionID != institutionID || quiz.CreatedBy != facultyID {
		return nil, ErrUnauthorizedQuiz
	}

	if quiz.IsPublished {
		return nil, ErrQuizAlreadyPublished
	}

	existingCount, err := s.materialRepo.CountByQuizID(ctx, quizID)
	if err != nil {
		return nil, err
	}

	if existingCount+int64(len(files)) > 10 {
		return nil, ErrMaterialLimitReached
	}

	
	const maxFileSize = 20 << 20
	var fileContentTypes []string

	for _, fileHeader := range files {
		if fileHeader.Size > maxFileSize {
			return nil, ErrFileTooLarge
		}

		f, err := fileHeader.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open uploaded file for validation: %w", err)
		}

		mtype, err := mimetype.DetectReader(f)
		f.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to detect MIME type: %w", err)
		}

		mimeStr := mtype.String()
		isAllowed := mimeStr == "application/pdf" ||
			mimeStr == "application/vnd.openxmlformats-officedocument.wordprocessingml.document" ||
			strings.HasPrefix(mimeStr, "image/")

		if !isAllowed {
			return nil, ErrUnsupportedFileType
		}

		fileContentTypes = append(fileContentTypes, mimeStr)
	}
	var writtenPaths []string
	var materials []domain.QuizMaterial

	for i, fileHeader := range files {
		cleanOrigName := filepath.Base(fileHeader.Filename)
		storedName := fmt.Sprintf("%s-%s", uuid.New().String(), cleanOrigName)
		storageKey := fmt.Sprintf("%s/%s/%s", institutionID.String(), quizID.String(), storedName)

		f, err := fileHeader.Open()
		if err != nil {
			s.cleanupPaths(ctx, writtenPaths)
			return nil, fmt.Errorf("failed to open file for storing: %w", err)
		}

		storedPath, err := s.storage.Store(ctx, storageKey, fileContentTypes[i], f)
		f.Close()
		if err != nil {
			s.cleanupPaths(ctx, writtenPaths)
			return nil, fmt.Errorf("storage write failed: %w", err)
		}

		writtenPaths = append(writtenPaths, storedPath)

		materials = append(materials, domain.QuizMaterial{
			QuizID:           quizID,
			UploadedBy:       facultyID,
			OriginalFilename: cleanOrigName,
			StoredFilename:   storedName,
			StoragePath:      storedPath,
			ContentType:      fileContentTypes[i],
			FileSize:         fileHeader.Size,
		})
	}
	if err := s.materialRepo.CreateBatch(ctx, materials); err != nil {
		s.cleanupPaths(ctx, writtenPaths)
		return nil, fmt.Errorf("failed to persist material metadata: %w", err)
	}

	logger.Info(ctx, "Quiz materials uploaded successfully", "quiz_id", quizID, "count", len(materials))

	var resp []dto.MaterialResponse
	for _, m := range materials {
		resp = append(resp, dto.MaterialResponse{
			ID:               m.ID,
			QuizID:           m.QuizID,
			UploadedBy:       m.UploadedBy,
			OriginalFilename: m.OriginalFilename,
			StoredFilename:   m.StoredFilename,
			ContentType:      m.ContentType,
			FileSize:         m.FileSize,
			CreatedAt:        m.CreatedAt,
		})
	}

	return &dto.UploadMaterialsResponse{Uploaded: resp}, nil
}

func (s *materialService) cleanupPaths(ctx context.Context, paths []string) {
	for _, p := range paths {
		if err := s.storage.Delete(ctx, p); err != nil {
			logger.Error(ctx, "Failed to clean up stored file during rollback", "path", p, "error", err)
		}
	}
}

func (s *materialService) ListMaterials(ctx context.Context, institutionID uuid.UUID, callerID uuid.UUID, callerRole domain.Role, quizID uuid.UUID) ([]dto.MaterialResponse, error) {
	quiz, err := s.quizRepo.GetQuizByID(ctx, quizID)
	if err != nil {
		return nil, err
	}

	if err := s.authorizeQuizAccess(ctx, quiz, institutionID, callerID, callerRole); err != nil {
		return nil, err
	}

	materials, err := s.materialRepo.GetByQuizID(ctx, quizID)
	if err != nil {
		return nil, err
	}

	var resp []dto.MaterialResponse
	for _, m := range materials {
		resp = append(resp, dto.MaterialResponse{
			ID:               m.ID,
			QuizID:           m.QuizID,
			UploadedBy:       m.UploadedBy,
			OriginalFilename: m.OriginalFilename,
			StoredFilename:   m.StoredFilename,
			ContentType:      m.ContentType,
			FileSize:         m.FileSize,
			CreatedAt:        m.CreatedAt,
		})
	}

	return resp, nil
}

func (s *materialService) DownloadMaterial(ctx context.Context, institutionID uuid.UUID, callerID uuid.UUID, callerRole domain.Role, quizID uuid.UUID, materialID uuid.UUID) (*dto.DownloadResult, error) {
	material, err := s.materialRepo.GetByID(ctx, materialID)
	if err != nil {
		return nil, err
	}

	if material.QuizID != quizID {
		return nil, ErrMaterialNotFound
	}

	quiz, err := s.quizRepo.GetQuizByID(ctx, quizID)
	if err != nil {
		return nil, err
	}

	if err := s.authorizeQuizAccess(ctx, quiz, institutionID, callerID, callerRole); err != nil {
		return nil, err
	}

	reader, err := s.storage.Open(ctx, material.StoragePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open material from storage: %w", err)
	}

	return &dto.DownloadResult{
		Reader:      reader,
		Filename:    material.OriginalFilename,
		ContentType: material.ContentType,
		Size:        material.FileSize,
	}, nil
}
