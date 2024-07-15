package controllers

type ResultSuccess struct {
	Result string `json:"result" binding:"required" example:"success"`
}

type ResultError struct {
	Error string `json:"error" binding:"required" example:"Some server error"`
}
