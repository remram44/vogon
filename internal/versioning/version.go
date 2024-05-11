package versioning

import (
	"fmt"
)

var Version = "v0.1"

func NameAndVersionString() string {
	return fmt.Sprintf("vogon %s", Version)
}
