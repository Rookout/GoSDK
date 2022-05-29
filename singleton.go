package rookout

import (
	"fmt"
)

func start(_ RookOptions) error {
	fmt.Println("Running Rookout in empty mode")
	return nil
}

func stop() error {
	return nil
}
