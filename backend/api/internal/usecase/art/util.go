package art

import (
	"fmt"
	"strings"
)

func requireTitle(title string) error {
	if strings.TrimSpace(title) == "" {
		return fmt.Errorf("title is required")
	}
	return nil
}
