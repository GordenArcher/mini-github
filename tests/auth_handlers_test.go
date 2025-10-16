package tests

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRegisterBadPayload(t *testing.T) {
	engine := gin.New()

	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer([]byte(`{"email":"x"}`)))
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	assert.Equal(t, 400, rec.Code)
}
