package types

type Publisher struct {
	TimeNow      int64  `json:"time_now"`
	AssignedName string `json:"assigned_name"`
	IPAddress    string `json:"ip_address"`
}
