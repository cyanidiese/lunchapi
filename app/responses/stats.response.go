package responses

type StatsResponse struct {
	Weight   float64 `json:"weight"`
	Calories float64 `json:"calories"`
	Price    float64 `json:"price"`
}
