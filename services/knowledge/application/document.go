// Package application contains the sample's backend-independent document model.
package application

// Document carries caller-generated embeddings; this example does not call a model.
type Document struct {
	ID, Text  string
	Embedding []float32
}
