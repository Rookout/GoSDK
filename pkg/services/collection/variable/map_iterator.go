package variable

import (
	"errors"
	"fmt"
	"github.com/Rookout/GoSDK/pkg/services/collection/memory"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/binary_info"
	"github.com/Rookout/GoSDK/pkg/services/instrumentation/dwarf/godwarf"
	"reflect"
)

type mapIterator struct {
	v          *Variable
	numbuckets uint64
	oldmask    uint64
	buckets    *Variable
	oldbuckets *Variable
	b          *Variable
	bidx       uint64

	tophashes *Variable
	keys      *Variable
	values    *Variable
	overflow  *Variable

	idx int64

	hashTophashEmptyOne uint64 
	hashMinTopHash      uint64 
}


func (v *Variable) mapIterator() (*mapIterator, error) {
	sv := v.clone()
	sv.RealType = resolveTypedef(&(sv.RealType.(*godwarf.MapType).TypedefType))
	sv = sv.MaybeDereference()
	v.Base = sv.Addr

	maptype, ok := sv.RealType.(*godwarf.StructType)
	if !ok {
		return nil, fmt.Errorf("wrong real type for map")
	}

	it := &mapIterator{v: v, bidx: 0, b: nil, idx: 0}

	if sv.Addr == 0 {
		it.numbuckets = 0
		return it, nil
	}

	for _, f := range maptype.Field {
		var err error
		field, _ := sv.toField(f)
		switch f.Name {
		case "count":
			v.Len, err = field.asInt()
		case "B":
			var b uint64
			b, err = field.asUint()
			it.numbuckets = 1 << b
			it.oldmask = (1 << (b - 1)) - 1
		case "buckets":
			it.buckets = field.MaybeDereference()
		case "oldbuckets":
			it.oldbuckets = field.MaybeDereference()
		}
		if err != nil {
			return nil, err
		}
	}

	if it.buckets.Kind != reflect.Struct || it.oldbuckets.Kind != reflect.Struct {
		return nil, errors.New("malformed map type: buckets, oldbuckets or overflow field not a struct")
	}

	it.hashTophashEmptyOne = hashTophashEmptyZero
	it.hashMinTopHash = hashMinTopHashGo111
	if binary_info.GoVersionAfterOrEqual(1, 12) {
		it.hashTophashEmptyOne = hashTophashEmptyOne
		it.hashMinTopHash = hashMinTopHashGo112
	}

	return it, nil
}

const (
	hashTophashEmptyZero = 0 
	hashTophashEmptyOne  = 1 
	hashMinTopHashGo111  = 4 
	hashMinTopHashGo112  = 5 
)

func (it *mapIterator) next() bool {
	for {
		if it.b == nil || it.idx >= it.tophashes.Len {
			r, _ := it.nextBucket()
			if !r {
				return false
			}
			it.idx = 0
		}
		tophash, _ := it.tophashes.sliceAccess(int(it.idx))
		h, err := tophash.asUint()
		if err != nil {
			it.v.Unreadable = fmt.Errorf("unreadable tophash: %v", err)
			return false
		}
		it.idx++
		if h != hashTophashEmptyZero && h != it.hashTophashEmptyOne {
			return true
		}
	}
}

func (it *mapIterator) key() *Variable {
	k, _ := it.keys.sliceAccess(int(it.idx - 1))
	return k
}

func (it *mapIterator) value() *Variable {
	v, _ := it.values.sliceAccess(int(it.idx - 1))
	return v
}

func (it *mapIterator) mapEvacuated(b *Variable) bool {
	if b.Addr == 0 {
		return true
	}
	for _, f := range b.DwarfType.(*godwarf.StructType).Field {
		if f.Name != "tophash" {
			continue
		}
		tophashes, _ := b.toField(f)
		tophash0var, _ := tophashes.sliceAccess(0)
		tophash0, err := tophash0var.asUint()
		if err != nil {
			return true
		}
		
		return tophash0 > it.hashTophashEmptyOne && tophash0 < it.hashMinTopHash
	}
	return true
}

func (v *Variable) sliceAccess(idx int) (*Variable, error) {
	wrong := false
	if v.Flags&VariableCPtr == 0 {
		wrong = idx < 0 || int64(idx) >= v.Len
	} else {
		wrong = idx < 0
	}
	if wrong {
		return nil, fmt.Errorf("index out of bounds")
	}
	mem := v.Mem
	if v.Kind != reflect.Array {
		mem = memory.DereferenceMemory(mem)
	}
	return v.spawn("", v.Base+uint64(int64(idx)*v.stride), v.fieldType, mem), nil
}

func (it *mapIterator) nextBucket() (bool, error) {
	if it.overflow != nil && it.overflow.Addr > 0 {
		it.b = it.overflow
	} else {
		it.b = nil

		for it.bidx < it.numbuckets {
			it.b = it.buckets.clone()
			it.b.Addr += uint64(it.buckets.DwarfType.Size()) * it.bidx

			if it.oldbuckets.Addr <= 0 {
				break
			}

			
			
			
			
			
			
			

			oldbidx := it.bidx & it.oldmask
			oldb := it.oldbuckets.clone()
			oldb.Addr += uint64(it.oldbuckets.DwarfType.Size()) * oldbidx

			if it.mapEvacuated(oldb) {
				break
			}

			if oldbidx == it.bidx {
				it.b = oldb
				break
			}

			
			
			it.b = nil
			it.bidx++
		}

		if it.b == nil {
			return false, nil
		}
		it.bidx++
	}

	if it.b.Addr <= 0 {
		return false, nil
	}

	it.b.Mem = memory.CacheMemory(it.b.Mem, it.b.Addr, int(it.b.RealType.Size()))

	it.tophashes = nil
	it.keys = nil
	it.values = nil
	it.overflow = nil

	for _, f := range it.b.DwarfType.(*godwarf.StructType).Field {
		field, err := it.b.toField(f)
		if err != nil {
			it.v.Unreadable = err
			return false, err
		}
		if field.Unreadable != nil {
			it.v.Unreadable = field.Unreadable
			return false, field.Unreadable
		}

		switch f.Name {
		case "tophash":
			it.tophashes = field
		case "keys":
			it.keys = field
		case "values":
			it.values = field
		case "overflow":
			it.overflow = field.MaybeDereference()
		}
	}

	
	if it.tophashes == nil || it.keys == nil || it.values == nil {
		it.v.Unreadable = fmt.Errorf("malformed map type")
		return false, it.v.Unreadable
	}

	if it.tophashes.Kind != reflect.Array || it.keys.Kind != reflect.Array || it.values.Kind != reflect.Array {
		it.v.Unreadable = errors.New("malformed map type: keys, values or tophash of a bucket is not an array")
		return false, it.v.Unreadable
	}

	if it.tophashes.Len != it.keys.Len {
		it.v.Unreadable = errors.New("malformed map type: inconsistent array length in bucket")
		return false, it.v.Unreadable
	}

	if it.values.fieldType.Size() > 0 && it.tophashes.Len != it.values.Len {
		
		
		it.v.Unreadable = errors.New("malformed map type: inconsistent array length in bucket")
		return false, it.v.Unreadable
	}

	if it.overflow.Kind != reflect.Struct {
		it.v.Unreadable = errors.New("malformed map type: buckets, oldbuckets or overflow field not a struct")
		return false, it.v.Unreadable
	}

	return true, nil
}
