package e2e

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func NewRandomShootName() string {
	length := 9
	id := uuid.New()
	result := strings.ReplaceAll(id.String(), "-", "")
	result = "p-" + result
	f := fmt.Sprintf("%%.%ds", length)
	result = fmt.Sprintf(f, result)
	return result
}
