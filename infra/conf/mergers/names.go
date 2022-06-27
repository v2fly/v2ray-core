package mergers

// GetAllNames get names of all formats
func GetAllNames() []string {
	names := make([]string, 0)
	for _, f := range mergersByName {
		names = append(names, f.Name)
	}
	return names
}
