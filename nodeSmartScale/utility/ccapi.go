package utility
import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"time"
)

var ApiStruct *Api

type Token struct {
	Access   string `json:"access_token"`
	Refresh  string `json:"refresh_token"`
	Expire   int64  `json:"expires_in"`
	Obtained time.Time
}

type Api struct {
	logEnabled     bool
	logBodyEnabled bool
	domain         string
	username       string
	password       string
	token          Token
}

func NewApi(dom string, user string, pass string, log bool, logBody bool) *Api {
	if dom == "" {
		panic("Need a DOMAIN to point the request somewhere...")
	}

	return &Api{domain: dom, username: user, password: pass, logEnabled: log, logBodyEnabled: logBody, token: Token{}}
}

func (api *Api) Put(url string, body *bytes.Buffer) (*http.Response, error) {
	u := fmt.Sprintf("http://api.%s%s", api.domain, url)
	// u := fmt.Sprintf("http://localhost:9999/%s/%s", api.domain, url)

	req, err := http.NewRequest("PUT", u, body)
	req.Body.Close()
	if err != nil {
		return nil, err
	}
	return api.doAuthenticatedRequest(req)
}

func (api *Api) Get(url string) (*http.Response, error) {
	u := fmt.Sprintf("http://api.%s%s", api.domain, url)
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	return api.doAuthenticatedRequest(req)
}

func (api *Api) GetFromLogSystem(url string) (*http.Response, error) {
	u := fmt.Sprintf("http://message.%s%s", api.domain, url)
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	return api.doAuthenticatedRequest(req)
}

func (api *Api) doAuthenticatedRequest(req *http.Request) (*http.Response, error) {
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+api.getAccessToken())
	req.Close = true
	//fmt.Println("token:" + api.getAccessToken())

	return api.doRequest(req)
}

func (api *Api) doRequest(req *http.Request) (*http.Response, error) {
	if api.logEnabled {
		reqString, _ := httputil.DumpRequest(req, api.logBodyEnabled)
		fmt.Printf("\n\n=====> Req: \n%s", string(reqString))
	}

	res, err := api.newClient().Do(req)
	//defer res.Body.Close()

	if api.logEnabled {
		resString, _ := httputil.DumpResponse(res, api.logBodyEnabled)
		if err != nil {
			fmt.Printf("-----> Resp ERROR: \n%s - %s", string(resString), err.Error())

		} else {
			fmt.Printf("-----> Resp: \n%s\n\n", string(resString))
		}
	}

	return res, err
}

func (api *Api) newClient() *http.Client {
	//Skip self signed cert validation...
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &http.Client{Transport: tr}
}

func (api *Api) getAccessToken() string {

	//If we have not token, get the first one
	if api.token.Access == "" {
		return api.getCleanToken()
	}

	since := time.Now().Unix() - api.token.Obtained.Unix()
	hasNotExpired := since < api.token.Expire-60*5   //提前5分钟，刷新token
	//hasNotExpired := since < 100

	if hasNotExpired {
		return api.token.Access
	}

	fmt.Println("[Token] Will refresh token")
	return api.getCleanToken()
}

func (api *Api) getCleanToken() string {

	fmt.Println("[Token] Will try to get a new token")
	url := fmt.Sprintf("https://uaa.%s/oauth/token?username=%s&password=%s&grant_type=password", api.domain, api.username, api.password)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		fmt.Printf("Could not try to get a new token: %s\n", err.Error())
		return ""
	}
	req.Header.Set("Authorization", "Basic Y2Y6") //Magic token! 'cf:' in base64
	//req.Header.Set("X-UAA-Endpoint","https://uaa.truepaas.com")
	req.Header.Set("Accept","application/json")
	req.Header.Set("Content-Type","application/x-www-form-urlencoded")
	res, err := api.doRequest(req)
	if err != nil {
		fmt.Printf("could not get a new token: %s\n", err.Error())
		return ""
	}
	json.NewDecoder(res.Body).Decode(&api.token)
	api.token.Obtained = time.Now()
	fmt.Printf("call cc api to obtain token,time now is %d\n",api.token.Obtained.Unix())
	return api.token.Access
}


//func (api *Api) refreshToken() string {
//
//	url := fmt.Sprintf("http://uaa.%s/oauth/token?grant_type=refresh_token&refresh_token=%s", api.domain, api.token.Refresh)
//	req, err := http.NewRequest("GET", url, nil)
//	if err != nil {
//		fmt.Printf("Could not try to refresh new token: %s\n", err.Error())
//		api.token.Access = ""
//		return ""
//	}
//
//	req.Header.Set("Authorization", "Basic Y2Y6") //Magic token! 'cf:' in base64
//	res, err := api.doRequest(req)
//	if err != nil {
//		fmt.Printf("could not refresh token: %s\n", err.Error())
//		api.token.Access = ""
//		return ""
//	}
//
//	json.NewDecoder(res.Body).Decode(&api.token)
//	api.token.Obtained = time.Now()
//	// fmt.Println("Token refreshed ok!!!!")
//	return api.token.Access
//}


