package requests

type ObjectCounter struct {
	Id    int64 `json:"id"`
	Count int64 `json:"count"`
}

type MenuUpdateRequest struct {
	DeliveryTime string          `json:"deliveryTime"`
	Deadline     string          `json:"deadline"`
	Items        []ObjectCounter `json:"dishes"`
}
