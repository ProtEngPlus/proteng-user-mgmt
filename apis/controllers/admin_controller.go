package controllers

import (
	"github.com/gin-gonic/gin"

	"github.com/protengplus/proteng-user-mgmt/models"
	"github.com/protengplus/proteng-user-mgmt/repositories"
	"github.com/protengplus/proteng-user-mgmt/utils/apiutil"
)

type AdminController struct {
	adminRepository repositories.AdminRepository
}

func NewAdminController(adminRepository repositories.AdminRepository) *AdminController {
	return &AdminController{adminRepository: adminRepository}
}

// GetAllAdmins retrieves all admins
func (uc *AdminController) GetAllAdmins(c *gin.Context) {
	admins, err := uc.adminRepository.GetAll()
	if err != nil {
		apiutil.ApiResponseInternalServerError(c, err)
		return
	}

	// Transform each admin using FilteredAdminResponse function
	var filteredAdmins []models.AdminResponse
	for _, admin := range admins {
		filteredAdmins = append(filteredAdmins, models.FilteredAdminResponse(admin))
	}

	apiutil.ApiResponseOk(c, filteredAdmins)
}

// GetAdmin retrieves a admin by ID
func (uc *AdminController) GetAdmin(c *gin.Context) {
	id := c.Param("id")
	admin, err := uc.adminRepository.FindById(id)
	if err != nil {
		apiutil.ApiResponseNotFound(c, err)
		return
	}

	apiutil.ApiResponseOk(c, models.FilteredAdminResponse(admin))
}

// CreateAdmin creates a new admin
func (uc *AdminController) CreateAdmin(c *gin.Context) {
	var admin models.Admin
	err := c.BindJSON(&admin)
	if err != nil {
		apiutil.ApiResponseErrorBadRequest(c, err, "error: invalid request body")
		return
	}

	err = uc.adminRepository.Create(&admin)
	if err != nil {
		apiutil.ApiResponseInternalServerError(c, err)
		return
	}

	apiutil.ApiResponseOk(c, models.FilteredAdminResponse(&admin))
}

// UpdateAdmin updates an existing admin
func (uc *AdminController) UpdateAdmin(c *gin.Context) {
	id := c.Param("id")
	admin, err := uc.adminRepository.FindById(id)
	if err != nil {
		apiutil.ApiResponseNotFound(c, err)
		return
	}
	err = c.BindJSON(&admin)
	if err != nil {
		apiutil.ApiResponseErrorBadRequest(c, err, "error: invalid request body")
		return
	}

	err = uc.adminRepository.Update(id, admin)
	if err != nil {
		apiutil.ApiResponseInternalServerError(c, err)
		return
	}

	apiutil.ApiResponseOk(c, models.FilteredAdminResponse(admin))
}

// DeleteAdmin deletes a admin by ID
func (uc *AdminController) DeleteAdmin(c *gin.Context) {
	id := c.Param("id")

	err := uc.adminRepository.Delete(id)
	if err != nil {
		apiutil.ApiResponseInternalServerError(c, err)
		return
	}

	apiutil.ApiResponseOk(c, nil)
}
