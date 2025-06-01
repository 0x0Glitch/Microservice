package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/0x0Glitch/toll-calculator/types"
)

type HTTPClient struct {
	Endpoint string
}

func NewHTTPClient(endpoint string) Client {
	return &HTTPClient{
		Endpoint: endpoint,
	}
}

func (c *HTTPClient) Aggregate(ctx context.Context,request types.AggregatorRequest) error {
	
	b, err := json.Marshal(&request)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", c.Endpoint+"/aggregate", bytes.NewReader(b))
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("the service responded with non 200 status code")
	}
	resp.Body.Close()
	return nil
}

func (c *HTTPClient) GetInvoice(ctx context.Context,id int) (*types.Invoice,error){
	invReq := types.GetInvoiceRequest{
		ObuID: uint64(id),
	}
	b ,err := json.Marshal(&invReq)
	if err != nil {
		return nil,err
	}
	req, err := http.NewRequest("POST", c.Endpoint+"/invoice", bytes.NewReader(b))
	if err != nil {
		return nil,err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil,err
	}
	

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("the service responded with non 200 status code")
	}
	var inv types.Invoice
	if err := json.NewDecoder(resp.Body).Decode(&inv); err != nil{
		return nil,err
	}
	defer resp.Body.Close()

	return &inv,nil

}
  