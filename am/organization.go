package am

// Organization represents an organization that has subscriped to our service
type Organization struct {
	OrgID           int32  `json:"org_id"`
	OrgName         string `json:"org_name"`
	OwnerEmail      string `json:"owner_email"`
	FirstName       string `json:"first_name"`
	LastName        string `json:"last_name"`
	Phone           string `json:"phone"`
	Country         string `json:"country"`
	StatePrefecture string `json:"state_prefecture"`
	Street          string `json:"street"`
	Address1        string `json:"address1"`
	Address2        string `json:"address2"`
	City            string `json:"city"`
	PostalCode      string `json:"postal_code"`
	CreationTime    int64  `json:"creation_time"`
	SubscriptionID  int    `json:"subscription_id"`
}
