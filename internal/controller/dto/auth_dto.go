package dto

type CredentialsRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

type FieldError struct {
	Location string `json:"location"`
	Param    string `json:"param"`
	Msg      string `json:"msg"`
}

type ValidationErrorResponse struct {
	Errors []FieldError `json:"errors"`
}

func NewValidationError(param, msg string) ValidationErrorResponse {
	return ValidationErrorResponse{
		Errors: []FieldError{
			{
				Location: "body",
				Param:    param,
				Msg:      msg,
			},
		},
	}
}
