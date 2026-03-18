package gateway

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	userpb "github.com/RAF-SI-2025/Banka-3-Backend/gen/user"
)

func SetupApi(router *gin.Engine, server *Server) {
	router.GET("/healthz", server.Healthz)
	router.Use(CORSMiddleware())
	api := router.Group("/api")
	{
		api.POST("/login", server.Login)
		api.POST("/logout", AuthenticatedMiddleware(server.UserClient), server.Logout)
		api.POST("/token/refresh", server.Refresh)
	}

	passwordReset := api.Group("/password-reset")
	{
		passwordReset.POST("/request", server.RequestPasswordReset)
		passwordReset.POST("/confirm", server.ConfirmPasswordReset)
	}

	clients := api.Group("/clients")
	{
		clients.POST("", server.CreateClientAccount)
	}

	employees := api.Group("/employees")
	{
		employees.POST("", server.CreateEmployeeAccount)
		employees.GET("/:id", server.GetEmployeeByID)
	}

	loans := api.Group("/loans")
	{
		loans.GET("", server.GetLoans)
		loans.GET("/:loanNumber", server.GetLoanByNumber)
	}

	api.POST("/loan-requests", server.CreateLoanRequest)
}

func (s *Server) Healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (s *Server) Logout(c *gin.Context) {
	email := c.GetString("email")
	println(email)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	_, err := s.UserClient.Logout(ctx, &userpb.LogoutRequest{
		Email: email,
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	c.Status(http.StatusAccepted)
}

func (s *Server) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBindError(c, err)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := s.UserClient.Login(ctx, &userpb.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  resp.AccessToken,
		"refresh_token": resp.RefreshToken,
	})
}

func (s *Server) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBindError(c, err)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := s.UserClient.Refresh(ctx, &userpb.RefreshRequest{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  resp.AccessToken,
		"refresh_token": resp.RefreshToken,
	})
}

func (s *Server) CreateClientAccount(c *gin.Context) {
	var req createClientAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBindError(c, err)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := s.UserClient.CreateClientAccount(ctx, &userpb.CreateClientRequest{
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		DateOfBirth: req.DateOfBirth,
		Gender:      req.Gender,
		Email:       req.Email,
		PhoneNumber: req.PhoneNumber,
		Address:     req.Address,
		Password:    req.Password,
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	if resp.Valid {
		c.JSON(http.StatusCreated, gin.H{
			"valid": true,
		})
		return
	}

	c.JSON(http.StatusUnprocessableEntity, gin.H{
		"valid": false,
	})
}

func (s *Server) CreateEmployeeAccount(c *gin.Context) {
	var req createEmployeeAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBindError(c, err)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := s.UserClient.CreateEmployeeAccount(ctx, &userpb.CreateEmployeeRequest{
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		DateOfBirth: req.DateOfBirth,
		Gender:      req.Gender,
		Email:       req.Email,
		PhoneNumber: req.PhoneNumber,
		Address:     req.Address,
		Username:    req.Username,
		Position:    req.Position,
		Department:  req.Department,
		Password:    req.Password,
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	if resp.Valid {
		c.JSON(http.StatusCreated, gin.H{
			"valid": true,
		})
		return
	}

	c.JSON(http.StatusUnprocessableEntity, gin.H{
		"valid": false,
	})
}

func (s *Server) GetEmployeeByID(c *gin.Context) {
	var uri getEmployeeByIDURI
	if err := c.ShouldBindUri(&uri); err != nil {
		c.String(http.StatusBadRequest, "employee id is required and must be a valid integer")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := s.UserClient.GetEmployeeById(ctx, &userpb.GetEmployeeByIdRequest{
		Id: uri.EmployeeID,
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         resp.Id,
		"first_name": resp.FirstName,
		"last_name":  resp.LastName,
		"email":      resp.Email,
		"position":   resp.Position,
		"active":     resp.Active,
	})
}

func (s *Server) RequestPasswordReset(c *gin.Context) {
	var req passwordResetRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBindError(c, err)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	_, err := s.UserClient.RequestPasswordReset(ctx, &userpb.PasswordActionRequest{
		Email: req.Email,
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "If that email exists, a reset link was sent.",
	})
}

func (s *Server) ConfirmPasswordReset(c *gin.Context) {
	var req passwordResetConfirmationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBindError(c, err)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	resp, err := s.UserClient.SetPasswordWithToken(ctx, &userpb.SetPasswordWithTokenRequest{
		Token:       req.Token,
		NewPassword: req.NewPassword,
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	if resp.Successful {
		c.Status(http.StatusOK)
	} else {
		c.Status(http.StatusUnprocessableEntity)
	}
}

func (s *Server) GetLoans(c *gin.Context) {
	var query getLoansQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		writeBindError(c, err)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := s.UserClient.GetLoans(ctx, &userpb.GetLoansRequest{
		LoanType:      query.LoanType,
		AccountNumber: query.AccountNumber,
		Status:        query.Status,
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	loans := make([]gin.H, 0, len(resp.Loans))
	for _, loan := range resp.Loans {
		loans = append(loans, gin.H{
			"loan_number":             loan.LoanNumber,
			"loan_type":               loan.LoanType,
			"account_number":          loan.AccountNumber,
			"loan_amount":             loan.LoanAmount,
			"repayment_period":        loan.RepaymentPeriod,
			"nominal_rate":            loan.NominalRate,
			"effective_rate":          loan.EffectiveRate,
			"agreement_date":          loan.AgreementDate,
			"maturity_date":           loan.MaturityDate,
			"next_installment_amount": loan.NextInstallmentAmount,
			"next_installment_date":   loan.NextInstallmentDate,
			"remaining_debt":          loan.RemainingDebt,
			"currency":                loan.Currency,
			"status":                  loan.Status,
		})
	}

	c.JSON(http.StatusOK, loans)
}

func (s *Server) GetLoanByNumber(c *gin.Context) {
	var uri getLoanByNumberURI
	if err := c.ShouldBindUri(&uri); err != nil {
		c.String(http.StatusBadRequest, "loan number is required")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := s.UserClient.GetLoanByNumber(ctx, &userpb.GetLoanByNumberRequest{
		LoanNumber: uri.LoanNumber,
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"loan_number":             resp.LoanNumber,
		"loan_type":               resp.LoanType,
		"account_number":          resp.AccountNumber,
		"loan_amount":             resp.LoanAmount,
		"repayment_period":        resp.RepaymentPeriod,
		"nominal_rate":            resp.NominalRate,
		"effective_rate":          resp.EffectiveRate,
		"agreement_date":          resp.AgreementDate,
		"maturity_date":           resp.MaturityDate,
		"next_installment_amount": resp.NextInstallmentAmount,
		"next_installment_date":   resp.NextInstallmentDate,
		"remaining_debt":          resp.RemainingDebt,
		"currency":                resp.Currency,
		"status":                  resp.Status,
	})
}

func (s *Server) CreateLoanRequest(c *gin.Context) {
	var req createLoanRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBindError(c, err)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	_, err := s.UserClient.CreateLoanRequest(ctx, &userpb.CreateLoanRequestRequest{
		AccountNumber:   req.AccountNumber,
		LoanType:        req.LoanType,
		Amount:          req.Amount,
		RepaymentPeriod: req.RepaymentPeriod,
		Currency:        req.Currency,
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	c.Status(http.StatusCreated)
}
