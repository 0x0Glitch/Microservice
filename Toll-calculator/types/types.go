package types

type OBUData struct {
	OBUID int     `json:"obuID"`
	Lat   float64 `json:"lat`
	Long  float64 `json:"long`
}

type Distance struct {
	Values  float64 `json:"value"`
	OBUID   int    `json:"obuID"`
	Unix   int64   `json:"unix"`
}

type Invoice struct{
	OBUID int `json:obuID`
	TotalDistance float64 `json:"totalDistance"`
	Amount float64 `json:"amount"`
}