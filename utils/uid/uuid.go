package uid

import (
	"encoding/base64"

	uuid "github.com/satori/go.uuid"
)

// NewID return a uuid
func NewID() string {
	id := uuid.NewV4()
	b64 := base64.URLEncoding.EncodeToString(id.Bytes()[:12])
	return b64
}
