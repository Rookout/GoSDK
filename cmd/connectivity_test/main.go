package main

import (
	"fmt"
	"github.com/Rookout/GoSDK/pkg"
	"github.com/Rookout/GoSDK/pkg/information"
	"os"
	"runtime"
)

func main() {
	fmt.Println("[Rookout] Testing connection to controller.")
	fmt.Printf("[Rookout] Rookout version: %s (%s)\n", information.VERSION, runtime.Version())

	err := startSingleton()
	if err != nil {
		fmt.Printf("[Rookout] Error occured during test: %v\n", err)
		fmt.Println("[Rookout] Test failed.")
		os.Exit(1)
	}

	fmt.Println("[Rookout] Test finished successfully.")
}

func startSingleton() error {
	s := pkg.GetSingleton()
	err := s.Start(&pkg.RookOptions{})
	if err != nil {
		return err
	}

	s.Stop()
	return nil
}
