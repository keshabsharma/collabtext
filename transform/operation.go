package transform

type Operation struct {
	// revision
	Revision uint64 `json:"revision"`

	ProcessedRevision uint64 `json:"processedrevision"`

	// operation -> insert or delete
	Op string `json:"op"`

	// position
	Position int32 `json:"position"`

	// string
	Str string `json:"str"`

	// origin
	Client string `json:"client"`

	// document name
	Document string `json:"document"`

	Error string `json:"error"`
}
