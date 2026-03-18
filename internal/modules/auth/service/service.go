package service

import (
	"context"

	"github.com/akza/akza-api/internal/domain"
	"github.com/akza/akza-api/internal/modules/auth/dto"
	"github.com/akza/akza-api/internal/pkg/apperror"
	jwtpkg "github.com/akza/akza-api/internal/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
)

type repo interface {
	FindByEmail(ctx context.Context, email string) (*domain.Admin, error)
	FindByID(ctx context.Context, id string) (*domain.Admin, error)
}

type Service struct {
	repo repo
	jwt  *jwtpkg.Manager
}

func New(repo repo, jwt *jwtpkg.Manager) *Service {
	return &Service{repo: repo, jwt: jwt}
}

func (s *Service) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	admin, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if err = bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.Password)); err != nil {
		return nil, apperror.ErrUnauthorized
	}
	token, err := s.jwt.GenerateToken(admin.ID, admin.Email)
	if err != nil {
		return nil, apperror.ErrInternal
	}
	return &dto.LoginResponse{
		AccessToken: token,
		Admin:       dto.AdminInfo{ID: admin.ID, Email: admin.Email, Name: admin.Name},
	}, nil
}

func (s *Service) Me(ctx context.Context, adminID string) (*dto.AdminInfo, error) {
	admin, err := s.repo.FindByID(ctx, adminID)
	if err != nil {
		return nil, err
	}
	return &dto.AdminInfo{ID: admin.ID, Email: admin.Email, Name: admin.Name}, nil
}
