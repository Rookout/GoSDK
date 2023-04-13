package locations_set

import (
	"sort"
	"sync"

	"github.com/Rookout/GoSDK/pkg/augs"
	"github.com/Rookout/GoSDK/pkg/augs/locations"
	"github.com/Rookout/GoSDK/pkg/types"
	"github.com/go-errors/errors"
	"github.com/gogo/protobuf/sortkeys"
)


type LocationsSet struct {
	*BreakpointStorage
	breakpoints     map[*augs.Breakpoint]map[types.AugID]locations.Location
	breakpointsLock sync.Mutex
}

func NewLocationsSet() *LocationsSet {
	return &LocationsSet{
		breakpoints:       make(map[*augs.Breakpoint]map[types.AugID]locations.Location),
		breakpointsLock:   sync.Mutex{},
		BreakpointStorage: newBreakpointStorage(),
	}
}

func (l *LocationsSet) Lock() {
	l.breakpointsLock.Lock()
}

func (l *LocationsSet) Unlock() {
	l.breakpointsLock.Unlock()
}


func (l *LocationsSet) AddLocation(location locations.Location, breakpoint *augs.Breakpoint) {
	l.Lock()
	defer l.Unlock()

	sort.Slice(breakpoint.Instances, func(i int, j int) bool {
		return breakpoint.Instances[i].Addr < breakpoint.Instances[j].Addr
	})
	if _, ok := l.breakpoints[breakpoint]; ok {
		l.breakpoints[breakpoint][location.GetAug().GetAugID()] = location
	} else {
		l.breakpoints[breakpoint] = map[types.AugID]locations.Location{location.GetAug().GetAugID(): location}
	}
}


func (l *LocationsSet) FindBreakpointByAddrs(addrs []uint64) (breakpoint *augs.Breakpoint, exists bool) {
	l.Lock()
	defer l.Unlock()

	sortkeys.Uint64s(addrs)
	for breakpoint := range l.breakpoints {
		if len(addrs) != len(breakpoint.Instances) {
			continue
		}

		isMatch := true
		for i, breakpointInstance := range breakpoint.Instances {
			if breakpointInstance.Addr != addrs[i] {
				isMatch = false
				break
			}
		}
		if isMatch {
			return breakpoint, true
		}
	}

	return nil, false
}


func (l *LocationsSet) FindBreakpointByAugID(augID types.AugID) (breakpoint *augs.Breakpoint, exists bool) {
	l.Lock()
	defer l.Unlock()

	return l.findBreakpointByAugID(augID)
}


func (l *LocationsSet) findBreakpointByAugID(augID types.AugID) (breakpoint *augs.Breakpoint, exists bool) {
	for bp, augIDs := range l.breakpoints {
		if _, exists := augIDs[augID]; exists {
			return bp, true
		}
	}

	return nil, false
}


func (l *LocationsSet) FindLocationsByBreakpointName(breakpointName string) (locations []locations.Location, exists bool) {
	l.Lock()
	defer l.Unlock()

	for bps, augIDsMap := range l.breakpoints {
		if bps.Name == breakpointName {
			for _, location := range augIDsMap {
				locations = append(locations, location)
			}
			return locations, true
		}
	}
	return nil, false
}



func (l *LocationsSet) ShouldClearBreakpoint(breakpoint *augs.Breakpoint) (bool, error) {
	l.Lock()
	defer l.Unlock()

	if augIDs, exists := l.breakpoints[breakpoint]; exists {
		return len(augIDs) == 0, nil
	}

	return false, errors.New("no such breakpoint")
}


func (l *LocationsSet) Breakpoints() (breakpoints []*augs.Breakpoint) {
	l.Lock()
	defer l.Unlock()

	for breakpoint := range l.breakpoints {
		breakpoints = append(breakpoints, breakpoint)
	}
	return breakpoints
}


func (l *LocationsSet) Locations() (locations []locations.Location) {
	l.Lock()
	defer l.Unlock()

	for _, locationsMap := range l.breakpoints {
		for _, location := range locationsMap {
			locations = append(locations, location)
		}
	}
	return locations
}



func (l *LocationsSet) RemoveLocation(augID types.AugID) {
	l.Lock()
	defer l.Unlock()

	if bp, exists := l.findBreakpointByAugID(augID); exists {
		_ = l.breakpoints[bp][augID].SetRemoved()
		delete(l.breakpoints[bp], augID)
	}
}


func (l *LocationsSet) RemoveBreakpoint(breakpoint *augs.Breakpoint) {
	l.Lock()
	defer l.Unlock()

	l.RemoveBreakpointUnsafe(breakpoint)
}



func (l *LocationsSet) RemoveBreakpointUnsafe(breakpoint *augs.Breakpoint) {
	delete(l.breakpoints, breakpoint)
}



func (l *LocationsSet) GetBreakpointsToRemoveUnsafe() []*augs.Breakpoint {
	var toClear []*augs.Breakpoint
	for bp := range l.breakpoints {
		if len(l.breakpoints[bp]) == 0 {
			toClear = append(toClear, bp)
		}
	}
	return toClear
}

func (l *LocationsSet) FindBreakpointByAddr(addr uint64) (*augs.BreakpointInstance, bool) {
	l.Lock()
	defer l.Unlock()

	for breakpoint := range l.breakpoints {
		matchingIndex := sort.Search(len(breakpoint.Instances), func(i int) bool {
			return breakpoint.Instances[i].Addr >= addr
		})
		if matchingIndex != len(breakpoint.Instances) && breakpoint.Instances[matchingIndex].Addr == addr {
			return breakpoint.Instances[matchingIndex], true
		}
	}

	return nil, false
}
