package domain

import (
	"fmt"
	"time"
)

const ReferenceCodePrefix = "REQ"

func NewReferenceCodePrefix() string {
	return fmt.Sprintf(`%s-%d`, ReferenceCodePrefix, time.Now().Year())
}