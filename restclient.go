/*
	@author: fersca, mlabarinas
*/

package restclient

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"sync"

	"github.com/hashicorp/golang-lru"
)

//Response is a struct that holds the information about the response of the call
type Response struct {
	Body          string
	Code          int
	Headers       map[string][]string
	CachedContent bool
	Staled        bool
}

//Rest Client (with cache) struct
type rClient struct {
	client  *http.Client
	baseURL string
	cache   *lru.Cache
	stale   bool
}

//PoolConfig is used to define a custom configuration for the pool
type PoolConfig struct {
	BaseURL             string
	MaxIdleConnsPerHost int
	Timeout             time.Duration
	Proxy               string
	CacheElements       int
	CacheState          bool
}

type Header struct {
	Key   string
	Value string
}

type NotFollowRedirectError struct{}

func (e *NotFollowRedirectError) Error() string {
	return "Don't follow redirect"
}

//Cache node
type cacheElement struct {
	Content string
	Headers map[string][]string
	Expires time.Time
}

//internal structure for mocks
type mockResponse struct {
	URL      string
	Regexp   bool
	Method   string
	Response Response
	Headers  map[string]string
	Body     string
}

//Connections pools
var pools = make(map[string]*rClient)

//List of mocks
var mocks []*mockResponse

//indicates if we have to use the mocks
var useMock bool

var notFollowRedirectError *NotFollowRedirectError = new(NotFollowRedirectError)

const (
	DEFAULT_MAX_IDLE_CONNECTIONS_PER_HOST = 100
)

//AddCustomPool create a new connection pool based on the sent parameters
func AddCustomPool(pattern string, config *PoolConfig) {
	//Create a transport for the connection
	transport := defaultTransport()

	//Add the max per host config
	if config.MaxIdleConnsPerHost > 0 {
		transport.MaxIdleConnsPerHost = config.MaxIdleConnsPerHost
	} else {
		transport.MaxIdleConnsPerHost = http.DefaultMaxIdleConnsPerHost
	}

	//Sets the proxy
	if config.Proxy != "" {
		proxyURL, _ := url.Parse(config.Proxy)
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	//Create the client
	client := &http.Client{Transport: transport}

	//Don't follow redirects
	client.CheckRedirect = func(request *http.Request, via []*http.Request) error {
		return notFollowRedirectError
	}

	//Sets the client timeout
	if config.Timeout != 0 {
		client.Timeout = config.Timeout * time.Millisecond
	}

	//Creates the client-cache struct
	rclient := new(rClient)
	rclient.client = client

	if config.BaseURL != "" {
		rclient.baseURL = config.BaseURL
	}

	//Create the cache if it was indicated
	if config.CacheElements > 0 {
		cache, _ := lru.New(config.CacheElements)
		rclient.cache = cache
		rclient.stale = config.CacheState
	}

	//save the pool
	pools[pattern] = rclient
}

//Get execute a HTTP GET call to the specified url using headers to forward
func Get(callURL string, headers ...Header) (*Response, error) {
	return performRequest(http.MethodGet, callURL, "", getHeadersMap(headers))
}

//Post execute a HTTP POST call to the specified url using headers to forward
func Post(callURL string, body string, headers ...Header) (*Response, error) {
	return performRequest(http.MethodPost, callURL, body, getHeadersMap(headers))
}

//Put execute a HTTP PUT call to the specified url using headers to forward
func Put(callURL string, body string, headers ...Header) (*Response, error) {
	return performRequest(http.MethodPut, callURL, body, getHeadersMap(headers))
}

//Delete execute a HTTP DELETE call to the specified url using headers to forward
func Delete(callURL string, headers ...Header) (*Response, error) {
	return performRequest(http.MethodDelete, callURL, "", getHeadersMap(headers))
}

//Head execute a HTTP HEAD call to the specified url using headers to forward
func Head(callURL string, headers ...Header) (*Response, error) {
	return performRequest(http.MethodHead, callURL, "", getHeadersMap(headers))
}

//Options execute a HTTP OPTIONS call to the specified url using headers to forward
func Options(callURL string, headers ...Header) (*Response, error) {
	return performRequest(http.MethodOptions, callURL, "", getHeadersMap(headers))
}

//AddMocks add seveal GET mocks for simple testing
func AddMocks(mocks map[string]Response) {
	//All all mocks in a bulk
	for key, value := range mocks {
		//fmt.Println("mock: ", key," - ", value)
		AddMock(key, http.MethodGet, "", value)
	}
}

var mutex = &sync.Mutex{}

//AddMock add a URL, headers and Response to the mock URL map
func AddMock(URL string, method string, body string, response Response, headers ...Header) {

	//If we are un production dont't load the mocks
	if !useMock {
		return
	}

	//Add the mock to the mock maps
	mResponse := new(mockResponse)
	mResponse.Response = response
	mResponse.Method = method
	mResponse.Headers = getHeadersMap(headers)
	mResponse.Body = body
	mResponse.URL = URL

	mutex.Lock()
	//Creste the map if doesnt exists of add the mock to the existing one
	if mocks == nil {
		mocks = make([]*mockResponse, 1)
		mocks[0] = mResponse
	} else {
		mocks = append(mocks, mResponse)
	}
	mutex.Unlock()

}

//URLasRegexp mark the URL to be evaluated as Regexp
func URLasRegexp(url string) {

	mutex.Lock()
	for _, mock := range mocks {
		if mock.URL == url {
			mock.Regexp = true
		}
	}
	mutex.Unlock()
}

//Clean mocks
func CleanMocks() {
	mutex.Lock()
	mocks = nil
	mutex.Unlock()
}

//DisableMock disable all the URL mocks
func DisableMock() {
	useMock = false
}

//UseMock inform if we are using mocks or not
func UseMock() bool {
	return useMock
}

//Execute the request
func performRequest(method string, callURL string, body string, headers map[string]string) (*Response, error) {
	//Get the rClient for the url
	rclient := getPool(callURL)

	if rclient.baseURL != "" && !strings.Contains(callURL, rclient.baseURL) {
		callURL = rclient.baseURL + callURL
	}

	//If theere is a mock for the url and we are in testing, return the mock response
	if useMock {
		r := searchMockCall(method, callURL, headers, body)
		if r != nil {
			return r, nil
		}
	}

	//Chech if we have to use the cache
	withCache := method == http.MethodGet && rclient.cache != nil

	var cachedResponse *Response

	if withCache {
		cachedResponse = getResponseFromCache(rclient, callURL)

		if cachedResponse != nil && !cachedResponse.Staled {
			return cachedResponse, nil
		}
	}

	var request *http.Request
	var error error

	//Create the request to the API
	if method == http.MethodPost || method == http.MethodPut {
		request, error = http.NewRequest(method, callURL, bytes.NewBuffer([]byte(body)))

	} else {
		request, error = http.NewRequest(method, callURL, nil)
	}

	//Checks for errors in the connection
	if error != nil {
		return nil, error
	}

	//Set headers
	setHeaders(request, headers)

	//perform the request through the client
	response, error := rclient.client.Do(request)

	//Defers the close of the response
	defer func() {
		//Check if it is not nil, because in case of error at opening gives an error
		if response != nil {
			response.Body.Close()
		}
	}()

	var rcResponse *Response

	isNotFollowRedirectError := false

	if urlError, ok := error.(*url.Error); ok && urlError.Err == notFollowRedirectError {
		isNotFollowRedirectError = true

		error = nil
	}

	if error != nil && !isNotFollowRedirectError {
		rcResponse = &Response{"", 0, nil, false, false}
	}

	var byteBody []byte

	if rcResponse == nil {
		//Read the response body
		if !isNotFollowRedirectError {
			byteBody, error = ioutil.ReadAll(response.Body)

			if error != nil {
				rcResponse = &Response{"", response.StatusCode, nil, false, false}
			}
		} else {
			byteBody = []byte("")
		}
	}

	if rcResponse == nil {
		rcResponse = &Response{string(byteBody), response.StatusCode, response.Header, false, false}
	}

	if withCache {
		//Chek if we got 200OK
		if rcResponse.Code == http.StatusOK {
			setResponseInCache(rclient, rcResponse, callURL)

		} else {
			//If we got some error and the state option is configured, return the last good cached response
			if rclient.stale && cachedResponse != nil {
				return cachedResponse, nil
			}
		}
	}

	return rcResponse, error
}

func defaultTransport() *http.Transport {
	//Create a transport for the connection
	transport := &http.Transport{
		DisableCompression: false,
		DisableKeepAlives:  false,
	}

	return transport
}

//InitDefaultPool initialize the rest client with the default settings
func initDefaultPool() *rClient {
	//Create a transport for the connection
	transport := defaultTransport()

	transport.MaxIdleConnsPerHost = DEFAULT_MAX_IDLE_CONNECTIONS_PER_HOST

	//create a http client to use, without timeout
	client := &http.Client{Transport: transport}

	client.CheckRedirect = func(request *http.Request, via []*http.Request) error {
		return notFollowRedirectError
	}

	//Creates the client-cache struct
	rclient := new(rClient)
	rclient.client = client

	pools["default"] = rclient

	return rclient
}

//Return the http client based on the URL to call
func getPool(callURL string) *rClient {
	//If we found a pool, return it
	for pattern, pool := range pools {
		if regexp.MustCompile(pattern).MatchString(callURL) {
			return pool
		}
	}

	//create a default pool
	pool := pools["default"]
	if pool == nil {
		pool = initDefaultPool()
	}

	return pool
}

//SetHeaders set the headers to the request
func setHeaders(request *http.Request, headers map[string]string) {
	//Set the headers to call the APIs
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Connection", "Keep-Alive")

	//set Content-Type
	if request.Method != http.MethodGet {
		request.Header.Set("Content-Type", "application/json")
	}

	//Forward the sent headers
	if headers != nil {
		for key, value := range headers {
			request.Header.Set(key, value)
		}
	}
}

//Check if the headers map is equal
func sameHeaders(h1 map[string]string, h2 map[string]string) bool {
	//Check both nil
	if h1 == nil && h2 == nil {
		return true
	}

	//Check one nil
	if h1 == nil || h2 == nil {
		return false
	}

	//Checl length
	if len(h1) != len(h2) {
		return false
	}

	//Check every value and key in the other map
	for key, value := range h1 {

		//Check if the key is in the other map
		val, ok := h2[key]
		if !ok {
			return false
		}

		//check the value
		if val != value {
			return false
		}

	}

	//If nothig was wrong return true (the maps are equal)
	return true
}

//Return the max age from the header
func getMaxAge(cacheControlValues []string) (int, error) {
	for _, value := range cacheControlValues {
		if strings.Contains(strings.ToLower(value), "max-age") {
			return strconv.Atoi(strings.Split(cacheControlValues[0], "=")[1])
		}
	}

	return 0, nil
}

//Chechs if have to go to the cache for the element
func getResponseFromCache(rclient *rClient, callURL string) *Response {
	//Chechs if it was previously saved
	if rclient.cache.Contains(callURL) {
		//Get the element from the cache
		element, _ := rclient.cache.Get(callURL)
		cacheElement := element.(*cacheElement)

		//If it is still valid, return the content from the cache
		if time.Now().Before(cacheElement.Expires) {
			return &Response{cacheElement.Content, 200, cacheElement.Headers, true, false}
		}

		//Save the expired response for staled calls
		if rclient.stale {
			return &Response{cacheElement.Content, 200, cacheElement.Headers, true, true}
		}
	}

	return nil
}

//Save the response to the cache
func setResponseInCache(rclient *rClient, response *Response, callURL string) {
	//Checks the cache control header
	cacheControl := response.Headers["Cache-Control"]

	if cacheControl != nil {
		//Check the max age value
		cacheControlValue, _ := getMaxAge(cacheControl)

		//Checks if we have to cache or not
		if cacheControlValue > 0 {
			//Create elemento to store the cache values
			cElement := new(cacheElement)
			cElement.Content = response.Body
			cElement.Headers = response.Headers
			cElement.Expires = time.Now().Add(time.Second * time.Duration(cacheControlValue))

			//Save the data in the cache
			rclient.cache.Add(callURL, cElement)
		}
	}
}

func getHeadersMap(headers []Header) map[string]string {
	headersMap := make(map[string]string)

	for _, header := range headers {
		headersMap[header.Key] = header.Value
	}

	return headersMap
}

//search for the url in the mocks array
func searchMockCall(method string, callURL string, headers map[string]string, body string) *Response {
	//Check every mock URL
	for _, mock := range mocks {

		var mached bool
		if mock.Regexp {
			mached, _ = regexp.MatchString(mock.URL, callURL)
		} else {
			mached = mock.URL == callURL
		}
		if mached && mock.Method == method && sameHeaders(headers, mock.Headers) && body == mock.Body {
			return &mock.Response
		}
	}

	return nil
}

//Init the Mock files
func init() {
	if os.Getenv("GO_ENVIRONMENT") != "production" {
		useMock = true
	}
}
