package geodbtools

// Writer defines a database writer instance
type Writer interface {
	// WriteDatabase writes the database given the database metadata and a given record tree
	WriteDatabase(meta Metadata, tree *RecordTree) (err error)
}
