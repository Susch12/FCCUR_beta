package api

import (
	"log"
	"net/http"

	"github.com/jesus/FCCUR/internal/models"
)

// withRoleRequired middleware ensures user has required role
func (s *Server) withRoleRequired(roles ...models.UserRole) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Get current user from JWT
			claims, err := s.getCurrentUser(r)
			if err != nil {
				respondJSON(w, http.StatusUnauthorized, map[string]string{
					"error": "Unauthorized - login required",
				})
				return
			}

			// Get full user from database
			user, err := s.db.GetUserByID(claims.UserID)
			if err != nil {
				log.Printf("Error getting user: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			// Check if user is active
			if !user.IsActive {
				respondJSON(w, http.StatusForbidden, map[string]string{
					"error": "Account is deactivated",
				})
				return
			}

			// Check if user has any of the required roles
			hasRole := false
			for _, role := range roles {
				if user.HasRole(role) {
					hasRole = true
					break
				}
			}

			if !hasRole {
				respondJSON(w, http.StatusForbidden, map[string]string{
					"error": "Insufficient permissions for this action",
				})
				return
			}

			next(w, r)
		}
	}
}

// withAdminOnly middleware ensures user is admin
func (s *Server) withAdminOnly(next http.HandlerFunc) http.HandlerFunc {
	return s.withRoleRequired(models.RoleAdmin)(next)
}

// withCanUpload middleware ensures user can upload
func (s *Server) withCanUpload(next http.HandlerFunc) http.HandlerFunc {
	return s.withRoleRequired(models.RoleAdmin, models.RoleProfessor)(next)
}

// withCanDelete middleware ensures user can delete
func (s *Server) withCanDelete(next http.HandlerFunc) http.HandlerFunc {
	return s.withRoleRequired(models.RoleAdmin)(next)
}

// checkCoursePermission checks if user can upload to a specific course
func (s *Server) checkCoursePermission(w http.ResponseWriter, r *http.Request, courseName string) bool {
	claims, err := s.getCurrentUser(r)
	if err != nil {
		return false
	}

	user, err := s.db.GetUserByID(claims.UserID)
	if err != nil {
		return false
	}

	return user.CanUploadToCourse(courseName)
}
