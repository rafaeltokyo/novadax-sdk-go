package novadax

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"time"
)

// APIClient - struct client
type APIClient struct {
	client *http.Client
	Env    string
	Token  string
}

// Error - struct error
type Error struct {
	Message string `json:"message"`
	Data    string `json:"data"`
}

// Data - struct response
type Data struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

//New - create a new client
func New(token, env string) *APIClient {
	return &APIClient{
		client: &http.Client{Timeout: 60 * time.Second},
		Env:    env,
		Token:  token,
	}
}

func (client *APIClient) Request(method, action string, body []byte, query interface{}, out interface{}) (error, *Error) {
	if client.client == nil {
		client.client = &http.Client{Timeout: 60 * time.Second}
	}
	endpoint := fmt.Sprintf("%s/%s", client.devProd(), action)
	req, err := http.NewRequest(method, endpoint, bytes.NewBuffer(body))
	if err != nil {
		return err, nil
	}
	req.Header.Add("Content-Type", "application/json")
	if query != nil {
		q := url.Values{}
		queryStruct := structToMap(query)
		for k, v := range queryStruct {
			q.Add(k, fmt.Sprintf("%v", v))
		}
		req.URL.RawQuery = q.Encode()
	}
	res, err := client.client.Do(req)
	if err != nil {
		return err, nil
	}
	bodyResponse, err := ioutil.ReadAll(res.Body)

	if res.StatusCode > 201 {
		var errAPI Error
		err = json.Unmarshal(bodyResponse, &errAPI)
		if err != nil {
			return err, nil
		}
		errAPI.Data = string(bodyResponse)
		return nil, &errAPI
	}
	response := Data{}
	err = json.Unmarshal(bodyResponse, &response)
	fmt.Printf("\n bodyResponse %s \n", bodyResponse)
	dataObjet, err := json.Marshal(response.Data)

	if err != nil {
		return err, nil
	}
	err = json.Unmarshal(dataObjet, out)
	fmt.Printf("\n err %s \n", err)
	if err != nil {
		return err, nil
	}
	return nil, nil
}

//devProd - check type Env
func (client *APIClient) devProd() string {
	if client.Env == "develop" {
		return "https://api.novadax.com/v1"
	}
	return "https://api.novadax.com/v1"
}

func structToMap(item interface{}) map[string]interface{} {

	res := map[string]interface{}{}
	if item == nil {
		return res
	}
	v := reflect.TypeOf(item)
	reflectValue := reflect.ValueOf(item)
	reflectValue = reflect.Indirect(reflectValue)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	for i := 0; i < v.NumField(); i++ {
		tag := v.Field(i).Tag.Get("json")
		field := reflectValue.Field(i).Interface()
		if tag != "" && tag != "-" {
			if v.Field(i).Type.Kind() == reflect.Struct {
				res[tag] = structToMap(field)
			} else {
				res[tag] = field
			}
		}
	}
	return res
}
