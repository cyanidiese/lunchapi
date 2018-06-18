package requests

type OrderingRequest struct {
	Items    []ObjectCounter   `json:"menuItems"`
}
