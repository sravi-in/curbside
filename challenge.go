package curbside

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

//const ChallengeBaseURL = "http://challenge.shopcurbside.com/"

const ChallengeBaseURL = "http://localhost:8000/"

type ChallengeRsp struct {
	Depth    int         `json:"depth"`
	ID       string      `json:"id"`
	Message  string      `json:"message"`
	Secret   string      `json:"secret"`
	ErrorMsg string      `json:"error"`
	Next     interface{} `json:"next"` // A []string or string
	Child    []string
}

func QueryServer(session, id string) (*ChallengeRsp, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", ChallengeBaseURL+id, nil)
	if err != nil {
		return nil, fmt.Errorf("new http request: %v", err)
	}

	req.Header.Add("Session", session)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get req session %q: %v", session, err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("reading rsp %q: %v", ChallengeBaseURL+id, err)
	}

	rsp, err := UnmarshalRsp(body)
	if err != nil {
		return nil,
			fmt.Errorf("parsing rsp %q session %q: %v", ChallengeBaseURL+id, session, err)
	}

	return rsp, nil
}

func UnmarshalRsp(body []byte) (*ChallengeRsp, error) {
	rsp := new(ChallengeRsp)
	if err := json.Unmarshal(body, rsp); err != nil {
		return nil, err
	}

	if rsp.ErrorMsg != "" {
		return nil, fmt.Errorf("error rsp received: %q", rsp.ErrorMsg)
	}

	if rsp.Next != nil {
		switch v := rsp.Next.(type) {
		case string:
			rsp.Child = append(rsp.Child, v)
		case []interface{}:
			for _, id := range v {
				v, ok := id.(string)
				if !ok {
					return nil,
						fmt.Errorf("unexpected type while unmarshaling next: %s",
							string(body))
				}
				rsp.Child = append(rsp.Child, v)
			}
		default:
			return nil, fmt.Errorf("unexpected type while unmarshaling next: %s", string(body))
		}

		// values other than string & []string ignored
	}

	return rsp, nil
}
