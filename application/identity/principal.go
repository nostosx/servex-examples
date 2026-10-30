// Package identity defines the application identity shared by authentication adapters.
package identity

type Principal struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func Valid(p Principal) bool { return p.ID != "" && len(p.ID) <= 512 && len(p.Name) <= 256 }
