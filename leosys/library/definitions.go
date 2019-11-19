package library

type UserResponse struct {
	Status string `json:"status"`
	Data   struct {
		ID                 int         `json:"id"`
		Enabled            bool        `json:"enabled"`
		Name               string      `json:"name"`
		Username           string      `json:"username"`
		Username2          interface{} `json:"username2"`
		Status             string      `json:"status"`
		LastLogin          string      `json:"lastLogin"`
		CheckedIn          bool        `json:"checkedIn"`
		LastIn             interface{} `json:"lastIn"`
		LastOut            interface{} `json:"lastOut"`
		LastInBuildingID   interface{} `json:"lastInBuildingId"`
		LastInBuildingName interface{} `json:"lastInBuildingName"`
		ViolationCount     int         `json:"violationCount"`
	} `json:"data"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

type TokenResponse struct {
	Status string `json:"status"`
	Data   struct {
		Token string `json:"token"`
	} `json:"data"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type CheckInResponse struct {
	Status string `json:"status"`
	Data   struct {
		ID       int    `json:"id"`
		Receipt  string `json:"receipt"`
		OnDate   string `json:"onDate"`
		Begin    string `json:"begin"`
		End      string `json:"end"`
		Location string `json:"location"`
	} `json:"data"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

type FreeBookResponse struct {
	Status string `json:"status"`
	Data   struct {
		ID        int    `json:"id"`
		Receipt   string `json:"receipt"`
		OnDate    string `json:"onDate"`
		Begin     string `json:"begin"`
		End       string `json:"end"`
		Location  string `json:"location"`
		CheckedIn bool   `json:"checkedIn"`
	} `json:"data"`
	Message string `json:"message"`
	Code    string `json:"code"`

}

type FiltersResponse struct {
	Status string `json:"status"`
	Data   struct {
		Buildings [][]interface{} `json:"buildings"`
		Rooms     [][]interface{} `json:"rooms"`
		Hours     int             `json:"hours"`
		Dates     []string        `json:"dates"`
	} `json:"data"`
	Message string `json:"message"`
	Code    string `json:"code"`
}