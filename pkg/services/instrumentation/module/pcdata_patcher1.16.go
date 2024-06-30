//go:build go1.16 && !go1.23
// +build go1.16,!go1.23

package module



func (p *PCDataPatcher) CreatePCFile(oldTable []PCDataEntry) ([]PCDataEntry, map[int32]int32, error) {
	newTable, err := p.createSanitized(oldTable,
		func(table []PCDataEntry) ([]PCDataEntry, error) {
			err := p.updateOffsets(table)
			return table, err
		})
	if err != nil {
		return nil, nil, err
	}
	return newTable, nil, nil
}
