package rookout

import (
	"fmt"
)

func start(_ RookOptions) error {
	fmt.Printf("Running Rookout in empty mode")
	return nil
}

func stop() error {
	return nil
}
