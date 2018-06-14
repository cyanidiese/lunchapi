package responses

type GeneralResponse struct {
	Name string `json:"name"`
	Message string `json:"message"`
	Status int64 `json:"status"`
	Data map[int64]int64 `json:"data"`
}
