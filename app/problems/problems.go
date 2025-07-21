package problems

type Document struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
}

type ValidationDocument struct {
	Document
	Errors map[string][]string `json:"errors,omitempty"`
}

func Validation(detail string, errors map[string][]string) ValidationDocument {
	return ValidationDocument{
		Document: Document{
			Type:   "https://datatracker.ietf.org/doc/html/rfc7231#section-6.5.1",
			Title:  "One or more validation errors occurred",
			Status: 400,
			Detail: detail,
		},
		Errors: errors,
	}
}

type ConflictDocument struct {
	Document
	Field string `json:"field"`
	Value string `json:"value,omitempty"`
}

func Conflict(field, value, detail string) ConflictDocument {
	return ConflictDocument{
		Document: Document{
			Type:   "https://datatracker.ietf.org/doc/html/rfc7231#section-6.5.8",
			Title:  "A resource already exists at that location",
			Status: 409,
			Detail: detail,
		},
		Field: field,
		Value: value,
	}
}
