//go:build !darwin
// +build !darwin

package safe_hook_installer

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/Rookout/GoSDK/pkg/logger"
	"github.com/Rookout/GoSDK/pkg/rookoutErrors"
)

type MemoryRegion struct {
	StartAddr   uint64
	EndAddr     uint64
	Permissions int
}

func parseMapsLine(line string) (*MemoryRegion, error) {
	fields := strings.Fields(line)
	if len(fields) < 5 {
		return nil, rookoutErrors.NewInvalidProcMapsLine(line)
	}

	addrFields := strings.Split(fields[0], "-")
	if len(addrFields) != 2 {
		return nil, rookoutErrors.NewInvalidProcMapsAddresses(line, fields[0])
	}
	startAddr, err := strconv.ParseUint(addrFields[0], 16, 64)
	if err != nil {
		return nil, rookoutErrors.NewInvalidProcMapsStartAddress(line, addrFields[0], err)
	}
	endAddr, err := strconv.ParseUint(addrFields[1], 16, 64)
	if err != nil {
		return nil, rookoutErrors.NewInvalidProcMapsEndAddress(line, addrFields[1], err)
	}

	var permissions int
	if strings.Contains(fields[1], "r") {
		permissions |= syscall.PROT_READ
	}
	if strings.Contains(fields[1], "w") {
		permissions |= syscall.PROT_WRITE
	}
	if strings.Contains(fields[1], "x") {
		permissions |= syscall.PROT_EXEC
	}

	return &MemoryRegion{
		StartAddr:   startAddr,
		EndAddr:     endAddr,
		Permissions: permissions,
	}, nil
}

func getCurrentMemoryProtection(addr uint64, size uint64) (int, rookoutErrors.RookoutError) {
	mapsFile, err := os.Open("/proc/self/maps")
	if err != nil {
		return 0, rookoutErrors.NewFailedToOpenProcMapsFile(err)
	}
	defer mapsFile.Close()
	startAddr := addr
	endAddr := addr + size

	var permissions int
	scanner := bufio.NewScanner(mapsFile)
	for scanner.Scan() {
		memoryRegion, err := parseMapsLine(scanner.Text())
		if err != nil {
			logger.Logger().WithError(err).Warning("Failed to parse maps line, continuing")
		}

		if startAddr < memoryRegion.EndAddr && memoryRegion.StartAddr < endAddr {
			permissions |= memoryRegion.Permissions
		} else if endAddr < memoryRegion.StartAddr {
			
			break
		}
	}
	return permissions, nil
}

func (h *HookWriter) changeHookPermissions(prot int) rookoutErrors.RookoutError {
	_, _, errno := syscall.Syscall(syscall.SYS_MPROTECT, uintptr(h.hookPageAlignedStart), uintptr(h.hookPageAlignedEnd-h.hookPageAlignedStart), uintptr(prot))
	if errno != 0 {
		return rookoutErrors.NewMprotectFailed(h.hookPageAlignedStart, int(h.hookPageAlignedEnd-h.hookPageAlignedStart), prot, errno.Error())
	}
	return nil
}
