package governance

func HasIndexEntry(index Index, id string) bool {
	for _, entry := range index.Entries {
		if entry.ID == id {
			return true
		}
	}
	return false
}

func MissingReferences(index Index, refs []string) []string {
	missing := make([]string, 0)
	for _, ref := range refs {
		if !HasIndexEntry(index, ref) {
			missing = append(missing, ref)
		}
	}
	return missing
}
