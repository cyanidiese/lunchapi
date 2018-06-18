package requests

type MenuUpdateRequest struct {
	DeliveryTime string          `json:"deliveryTime"`
	Deadline     string          `json:"deadline"`
	Items        []ObjectCounter `json:"dishes"`
}
