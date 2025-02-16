package services

import (
	"errors"
	"fmt"
	"log"
	"neobase-ai/config"
	"neobase-ai/internal/apis/dtos"
	"neobase-ai/internal/models"
	"neobase-ai/internal/repositories"
	"neobase-ai/internal/utils"
	"net/http"
	"time"
)

type AuthService interface {
	Signup(req *dtos.SignupRequest) (*dtos.AuthResponse, uint, error)
	Login(req *dtos.LoginRequest) (*dtos.AuthResponse, uint, error)
	GenerateUserSignupSecret(req *dtos.UserSignupSecretRequest) (*models.UserSignupSecret, uint, error)
	RefreshToken(refreshToken string) (*dtos.RefreshTokenResponse, uint32, error)
	Logout(refreshToken string, accessToken string) (uint32, error)
}

type authService struct {
	userRepo   repositories.UserRepository
	jwtService utils.JWTService
	tokenRepo  repositories.TokenRepository
}

func NewAuthService(userRepo repositories.UserRepository, jwtService utils.JWTService, tokenRepo repositories.TokenRepository) AuthService {
	return &authService{
		userRepo:   userRepo,
		jwtService: jwtService,
		tokenRepo:  tokenRepo,
	}
}

func (s *authService) Signup(req *dtos.SignupRequest) (*dtos.AuthResponse, uint, error) {
	// Check if user exists

	if req.Username == config.Env.AdminUser {
		return nil, http.StatusBadRequest, errors.New("username already exists")
	}

	validUserSignupSecret := s.userRepo.ValidateUserSignupSecret(req.UserSignupSecret)
	if !validUserSignupSecret {
		return nil, http.StatusUnauthorized, errors.New("invalid user signup secret")
	}
	existingUser, err := s.userRepo.FindByUsername(req.Username)
	if err != nil {
		return nil, http.StatusNotFound, err
	}
	if existingUser != nil {
		return nil, http.StatusBadRequest, errors.New("username already exists")
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	// Create user
	user := &models.User{
		Username: req.Username,
		Password: hashedPassword,
		Base: models.Base{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, http.StatusBadRequest, err
	}

	// Generate token
	accessToken, err := s.jwtService.GenerateToken(user.ID.Hex())
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(user.ID.Hex())
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	err = s.tokenRepo.StoreRefreshToken(user.ID.Hex(), *refreshToken)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	go func() {
		err := s.userRepo.DeleteUserSignupSecret(req.UserSignupSecret)
		if err != nil {
			log.Println("failed to delete user signup secret:" + err.Error())
		}
	}()

	return &dtos.AuthResponse{
		AccessToken:  *accessToken,
		RefreshToken: *refreshToken,
		User:         *user,
	}, http.StatusCreated, nil
}

func (s *authService) Login(req *dtos.LoginRequest) (*dtos.AuthResponse, uint, error) {
	var authUser *models.User
	var err error
	// Check if it's Admin User
	if req.Username == config.Env.AdminUser {
		log.Println("Admin User Login")
		if req.Password != config.Env.AdminPassword {
			return nil, http.StatusUnauthorized, errors.New("invalid password")
		}
		user, err := s.userRepo.FindByUsername(req.Username)
		// Checking if Admin user exists in the DB, if not then create user for admin creds
		if err != nil || user == nil {
			log.Println("Admin User not found, creating user")
			// Hash password
			hashedPassword, err := utils.HashPassword(req.Password)
			if err != nil {
				return nil, http.StatusBadRequest, err
			}

			// Create user
			authUser = &models.User{
				Username: req.Username,
				Password: hashedPassword,
				Base: models.Base{
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			}

			if err = s.userRepo.Create(authUser); err != nil {
				log.Println("Failed to create admin user:" + err.Error())
				return nil, http.StatusBadRequest, err
			}
		}
	} else {
		log.Println("Non-Admin User Login")
		authUser, err = s.userRepo.FindByUsername(req.Username)
		if err != nil {
			log.Println("Failed to find user:" + err.Error())
			return nil, http.StatusNotFound, err
		}
		if authUser == nil {
			log.Println("User not found")
			return nil, http.StatusUnauthorized, errors.New("invalid credentials")
		}

		if !utils.CheckPasswordHash(req.Password, authUser.Password) {
			log.Println("Invalid credentials")
			return nil, http.StatusUnauthorized, errors.New("invalid credentials")
		}
	}
	accessToken, err := s.jwtService.GenerateToken(authUser.ID.Hex())
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(authUser.ID.Hex())
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	err = s.tokenRepo.StoreRefreshToken(authUser.ID.Hex(), *refreshToken)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	return &dtos.AuthResponse{
		AccessToken:  *accessToken,
		RefreshToken: *refreshToken,
		User:         *authUser,
	}, http.StatusOK, nil
}

func (s *authService) GenerateUserSignupSecret(req *dtos.UserSignupSecretRequest) (*models.UserSignupSecret, uint, error) {
	if req.Username != config.Env.AdminUser || req.Password != config.Env.AdminPassword {
		return nil, http.StatusUnauthorized, errors.New("invalid credentials for the admin")
	}

	secret := utils.GenerateSecret()

	createdSecret, err := s.userRepo.CreateUserSignUpSecret(secret)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	return createdSecret, http.StatusCreated, nil
}

func (s *authService) RefreshToken(refreshToken string) (*dtos.RefreshTokenResponse, uint32, error) {
	// Validate the refresh token
	claims, err := s.jwtService.ValidateToken(refreshToken)
	if err != nil {
		return nil, http.StatusUnauthorized, fmt.Errorf("invalid refresh token")
	}

	log.Println("Validating refresh token:", refreshToken)
	// Check if the refresh token exists in Redis
	if !s.tokenRepo.ValidateRefreshToken(*claims, refreshToken) {
		return nil, http.StatusUnauthorized, fmt.Errorf("refresh token not found")
	}

	// Generate new tokens
	accessToken, err := s.jwtService.GenerateToken(*claims)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return &dtos.RefreshTokenResponse{
		AccessToken: *accessToken,
	}, http.StatusOK, nil
}

func (s *authService) Logout(refreshToken string, accessToken string) (uint32, error) {
	// Validate the refresh token
	claims, err := s.jwtService.ValidateToken(refreshToken)
	if err != nil {
		return http.StatusUnauthorized, fmt.Errorf("invalid refresh token")
	}

	// Delete the refresh token from Redis
	if err := s.tokenRepo.DeleteRefreshToken(*claims, refreshToken); err != nil {
		return http.StatusInternalServerError, err
	}

	// Blacklist the access token until its original expiration
	_, err = s.jwtService.ValidateToken(accessToken)
	if err != nil {
		return http.StatusUnauthorized, fmt.Errorf("invalid access token")
	}

	if err := s.tokenRepo.BlacklistToken(accessToken, time.Duration(config.Env.JWTExpirationMilliseconds)); err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}
