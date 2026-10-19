package models

type DogsRes struct {
	Name  string `json:"name"`
	DogID int    `json:"dog_id"`
	Type  string `json:"type"`
}

type ResultData struct {
	Data       []DogsRes `json:"data"`
	Name       string    `json:"name"`
	Count      int       `json:"count"`
	SumRed     int       `json:"sum_red"`
	SumGreen   int       `json:"sum_green"`
	SumPink    int       `json:"sum_pink"`
	SumNoColor int       `json:"sum_nocolor"`
}
