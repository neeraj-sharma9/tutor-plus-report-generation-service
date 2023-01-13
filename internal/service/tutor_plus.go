package service

import (
	"encoding/json"
	"fmt"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/config"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/constant"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/contract"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/helper"
	"github.com/neeraj-sharma9/tutor-plus-report-generation-service/internal/logger"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

type TutorPlusService struct {
	config *config.Config
}

func TutorPlusServiceInitializer(config *config.Config) *TutorPlusService {
	return &TutorPlusService{config: config}
}

func (tPS *TutorPlusService) GetUserDetails(mprReq *helper.MPRReq) contract.UserDetailsResponse {
	baseUrl := tPS.config.TutorPlusBaseURL
	params := make(map[string]string)
	params["user_id"] = strconv.FormatInt(mprReq.UserId, 10)
	if mprReq.EpchFrmDate > 0 {
		params["start_date"] = time.Unix(mprReq.EpchFrmDate, 0).Format(constant.TIME_LAYOUT)
	}
	if mprReq.EpchToDate > 0 {
		params["end_date"] = time.Unix(mprReq.EpchToDate, 0).Format(constant.TIME_LAYOUT)
	}
	baseUrl += "/internal_api/neo/progress_report/mt_user_class_detail"
	apiResponse, err := tPS.GetApiExecutor(baseUrl, params)
	var userDetails contract.UserDetailsResponse
	if string(apiResponse) == "{\"data\":{}}" {
		mprReq.ReqStatus = false
		mprReq.State = constant.JSON_GENERATION_FAILED
		mprReq.ErrorMsg = "Insufficient data from Neo classes API"
		return userDetails
	}
	if err != nil {
		mprReq.ReqStatus = false
		mprReq.State = constant.TUTOR_PLUS_DATA_API_5XX_ERROR
		return userDetails
	}
	err = json.Unmarshal(apiResponse, &userDetails)
	if err != nil {
		logger.Log.Sugar().Errorf("Error unmarshalling userDetails data: %v, response: %+v,"+
			" baseURL: %v, params: %+v", err, apiResponse, baseUrl, params)
	}
	return userDetails
}

func (tPS *TutorPlusService) GetApiExecutor(baseURL string, params map[string]string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", baseURL, nil)
	if err != nil {
		logger.Log.Sugar().Errorf("Error creating new http request error: %v, for request: %v and params: %v",
			err, baseURL, params)
		return []byte(`{}`), fmt.Errorf(constant.JSON_GENERATION_FAILED)
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("auth_key", tPS.config.APIKey)
	req.Header.Add("auth_secret", tPS.config.APISecret)
	query := req.URL.Query()
	if len(params) > 0 {

		for key, value := range params {
			query.Add(key, value)
		}
	}
	query.Add("api_key", tPS.config.APIKey)
	query.Add("api_secret", tPS.config.APISecret)

	req.URL.RawQuery = query.Encode()
	resp, err := client.Do(req)
	if err != nil {
		logger.Log.Sugar().Errorf("Error sending request to tutor+ error: %v, for request: %v, params: %+v, query: %v",
			err, baseURL, params, query)
		return []byte(`{}`), fmt.Errorf(constant.JSON_GENERATION_FAILED)
	}

	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Log.Sugar().Errorf("Error reading tutor+ request response error: %v, for request: %v, params: %+v,"+
			" query: %v", err, baseURL, params, query)
		return []byte(`{}`), fmt.Errorf(constant.JSON_GENERATION_FAILED)
	}

	if resp.StatusCode >= 400 && resp.StatusCode <= 499 {
		logger.Log.Sugar().Errorf("tutor+ api 4xx error with resp: %+v, for request: %v, params: %+v,"+
			" query: %v", resp, baseURL, params, query)
		return []byte(`{}`), fmt.Errorf(constant.JSON_GENERATION_FAILED)
	}

	if resp.StatusCode >= 500 && resp.StatusCode <= 599 {
		logger.Log.Sugar().Errorf("tutor+ api 5xx error with resp: %+v, for request: %v, params: %+v,"+
			" query: %v", resp, baseURL, params, query)
		return []byte(`{}`), fmt.Errorf(constant.JSON_GENERATION_FAILED)
	}

	return bodyBytes, nil
}
