//go:build go1.15 && !go1.16
// +build go1.15,!go1.16

package module



func (p *PCDataPatcher) CreatePCFile(oldTable []PCDataEntry) ([]PCDataEntry, map[int32]int32, error) {
	oldFilenosToNew := make(map[int32]int32)
	i := int32(0)
	for _, entry := range oldTable {
		if _, ok := oldFilenosToNew[entry.Value]; !ok {
			oldFilenosToNew[entry.Value] = i
			i++
		}
	}

	newTable, err := p.createSanitized(oldTable,
		func(table []PCDataEntry) ([]PCDataEntry, error) {
			err := p.updateOffsets(table)
			if err != nil {
				return nil, err
			}

			
			for i := range table {
				table[i].Value = oldFilenosToNew[table[i].Value]
			}
			return table, nil
		})
	if err != nil {
		return nil, nil, err
	}

	return newTable, oldFilenosToNew, nil
}
