package requests

type UpdateProfileRequest struct {
	FirstName  string `json:"firstName"`
	LastName   string `json:"lastName"`
	Alias      string `json:"alias" description:"Identifier for providers to put on the top of food box"`
	Password   string `json:"password"`
	ProviderId int64  `json:"providerId"`
	OfficeId   int64  `json:"officeId"`
	ImageGuid  string `json:"imageGuid"`
	Timezone   string `json:"timezone"`
	Language   string `json:"lang"`
}
