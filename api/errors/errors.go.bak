package errors

import (
	"encoding/json"
	"net/http"
)

type APIError interface {
	ID() string
	Message() string
	Code() int
	Data() interface{}
	error
}

func (s apiError) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"code":    s.ID(),
		"message": s.Message(),
	})
}

// Errors that implement the PublicErrorer interface have different
// errors for internal and public consumption.
type PublicErrorer interface {
	PublicError() APIError
	APIError
}

type apiError struct {
	id      string `json:"id"`
	message string `json:"message"`
	code    int
}

func New(id string, message string, code int) APIError {
	return &apiError{id, message, code}
}

func (err *apiError) Error() string {
	return err.Message()
}

func (err *apiError) ID() string {
	return err.id
}

func (err *apiError) Message() string {
	return err.message
}

func (err *apiError) Code() int {
	return err.code
}

func (err *apiError) Data() interface{} {
	return nil
}

type publicError struct {
	*apiError
	public APIError
}

func NewPublic(message string, public APIError) PublicErrorer {
	return &publicError{New("internal_error", message, 500).(*apiError), public}
}

func (err *publicError) PublicError() APIError {
	return err.public
}

func (err *publicError) Error() string {
	return err.message
}

type ValidationField struct {
	Name     string `json:"name"`
	ReasonID string `json:"reason_id"`
	Reason   string `json:"reason"`
}

var _ APIError = &ValidationError{}

type ValidationError struct {
	error
	Fields []ValidationField
}

func (err *ValidationError) ID() string {
	return "failed_validation"
}

func (err *ValidationError) Message() string {
	return "Some fields failed validation"
}

func (err *ValidationError) Code() int {
	return http.StatusBadRequest
}

func (err *ValidationError) Data() interface{} {
	return err.Fields
}

func (err *ValidationError) Error() string {
	return "failed validation"
}

func (err *ValidationError) Add(name, id, reason string) {
	err.Fields = append(err.Fields, ValidationField{name, id, reason})
}

func (err *ValidationError) Valid() bool {
	return len(err.Fields) == 0
}
