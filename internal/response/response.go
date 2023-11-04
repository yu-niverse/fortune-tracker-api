package response

type Response struct {
	Status  bool        `json:"Status"`
	Data    interface{} `json:"Data"`
	Message interface{} `json:"Message"`
}

type RegisterResponse struct {
	UUID string `json:"UUID"`
	Email string `json:"Email"`
}

type LoginResponse struct {
	UUID string `json:"UUID"`
	Token string `json:"Token"`
}

type ULIDResponse struct {
	ULID string `json:"ULID"`
}

type UTIDResponse struct {
	UTID string `json:"UTID"`
}

func New() *Response {
	return &Response{
		Status:  false,
		Data:    nil, 
		Message: nil,
	}
}
