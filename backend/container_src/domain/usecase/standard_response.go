package usecase

type StandardResponse struct {
	Status  string      `json:"status"`
	Code    int         `json:"int"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}
