package main

import (
	"encoding/json"
	"net/http"
	"strconv"
)

// DataResp response data http
type DataResp struct {
	Data       interface{} `json:"data,omitempty"`
	IsSuccess  bool        `json:"is_success"`
	ErrMessage string      `json:"error_message,omitempty"`
}

// GetConfigHandler function to get config now
func (handler *HTTPHandler) GetConfigHandler(w http.ResponseWriter, r *http.Request) {
	resp := DataResp{}
	defer func() {
		responseByte, _ := json.Marshal(resp)
		w.Header().Set("Content-Type", "application/json")
		w.Write(responseByte)
	}()
	config, err := handler.dbkvlib.GetConfig(r.FormValue("config_key"))
	if err != nil {
		resp.ErrMessage = err.Error()
		return
	}
	resp.Data = config
	resp.IsSuccess = true
}

// UpdateConfigHandler function to update config and reinit config
func (handler *HTTPHandler) UpdateConfigHandler(w http.ResponseWriter, r *http.Request) {
	resp := DataResp{}
	defer func() {
		responseByte, _ := json.Marshal(resp)
		w.Header().Set("Content-Type", "application/json")
		w.Write(responseByte)
	}()
	err := handler.dbkvlib.UpdateConfig(r.FormValue("config_key"), r.FormValue("config_value"))
	if err != nil {
		resp.ErrMessage = err.Error()
		return
	}
	resp.IsSuccess = true
}

// InsertConfigHandler function to insert config
func (handler *HTTPHandler) InsertConfigHandler(w http.ResponseWriter, r *http.Request) {
	resp := DataResp{}
	defer func() {
		responseByte, _ := json.Marshal(resp)
		w.Header().Set("Content-Type", "application/json")
		w.Write(responseByte)
	}()
	err := handler.dbkvlib.InsertConfig(r.FormValue("config_key"), r.FormValue("config_value"))
	if err != nil {
		resp.ErrMessage = err.Error()
		return
	}
	resp.IsSuccess = true
}

// CheckEligibleUserHandler testing handler config
func (handler *HTTPHandler) CheckEligibleUserHandler(w http.ResponseWriter, r *http.Request) {
	resp := DataResp{}

	defer func() {
		responseByte, _ := json.Marshal(resp)
		w.Header().Set("Content-Type", "application/json")
		w.Write(responseByte)
	}()
	age, err := strconv.Atoi(r.FormValue("age"))
	if err != nil {
		resp.ErrMessage = err.Error()
		return
	}

	dataCfg, err := handler.dbkvlib.GetConfig("key1")
	if err != nil {
		resp.ErrMessage = err.Error()
		return
	}

	dataUserCfg, ok := dataCfg.(*UserConfig)
	if !ok {
		resp.ErrMessage = "Try Again"
		return
	}

	type RespEligible struct {
		IsEligible bool `json:"is_eligible"`
	}

	resp.IsSuccess = true

	resp.Data = RespEligible{
		IsEligible: age >= dataUserCfg.MinAge && age <= dataUserCfg.MaxAge,
	}
}

// CheckMethodHandler handler to provide user must using new method or not
func (handler *HTTPHandler) CheckMethodHandler(w http.ResponseWriter, r *http.Request) {
	resp := DataResp{}

	defer func() {
		responseByte, _ := json.Marshal(resp)
		w.Header().Set("Content-Type", "application/json")
		w.Write(responseByte)
	}()
	dataCfg, err := handler.dbkvlib.GetConfig("key4")
	if err != nil {
		resp.ErrMessage = err.Error()
		return
	}
	useNewMethod, ok := dataCfg.(bool)
	if !ok {
		resp.ErrMessage = "Try Again"
		return
	}

	type RespDecideMethod struct {
		UseNewMethod bool `json:"use_new_method"`
	}

	resp.IsSuccess = true
	resp.Data = RespDecideMethod{
		UseNewMethod: useNewMethod,
	}

}
