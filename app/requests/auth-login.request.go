package requests

type AuthLoginRequest struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	GrantType    string `json:"grant_type"`
	AuctionGuid  string `json:"auctionGuid"`
}
