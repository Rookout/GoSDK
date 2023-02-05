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
	breakpoints     map[*augs.Breakpoint]map[types.AugId]locations.Location
	breakpointsLock sync.Mutex
}

func NewLocationsSet() *LocationsSet {
	return &LocationsSet{
		breakpoints:       make(map[*augs.Breakpoint]map[types.AugId]locations.Location),
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
		l.breakpoints[breakpoint][location.GetAug().GetAugId()] = location
	} else {
		l.breakpoints[breakpoint] = map[types.AugId]locations.Location{location.GetAug().GetAugId(): location}
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


func (l *LocationsSet) FindBreakpointByAugId(augId types.AugId) (breakpoint *augs.Breakpoint, exists bool) {
	l.Lock()
	defer l.Unlock()

	return l.findBreakpointByAugId(augId)
}


func (l *LocationsSet) findBreakpointByAugId(augId types.AugId) (breakpoint *augs.Breakpoint, exists bool) {
	for bp, augIds := range l.breakpoints {
		if _, exists := augIds[augId]; exists {
			return bp, true
		}
	}

	return nil, false
}


func (l *LocationsSet) FindLocationsByBreakpointName(breakpointName string) (locations []locations.Location, exists bool) {
	l.Lock()
	defer l.Unlock()

	for bps, augIdsMap := range l.breakpoints {
		if bps.Name == breakpointName {
			for _, location := range augIdsMap {
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

	if augIds, exists := l.breakpoints[breakpoint]; exists {
		return len(augIds) == 0, nil
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



func (l *LocationsSet) RemoveLocation(augId types.AugId) {
	l.Lock()
	defer l.Unlock()

	if bp, exists := l.findBreakpointByAugId(augId); exists {
		_ = l.breakpoints[bp][augId].SetRemoved()
		delete(l.breakpoints[bp], augId)
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
