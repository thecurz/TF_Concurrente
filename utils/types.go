package utils

type ServerMessage struct {
	Message   string   `json:"message,omitempty"`
	Partition []Review `json:"partition,omitempty"`
}

type Review struct {
	ReviewID        string  `json:"review_id"`
	ProductID       string  `json:"product_id"`
	ReviewerID      string  `json:"reviewer_id"`
	Stars           float64 `json:"stars"`
	ProductCategory string  `json:"product_category"`
	// Add other fields if needed
}

type ResultData struct {
	Recommendations map[string][]string `json:"recommendations"` // Map of user IDs to recommended product IDs
}
