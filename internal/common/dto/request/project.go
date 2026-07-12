package request

// CreateProjectReq is the user-submitted input used to create a project.
type CreateProjectReq struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}
