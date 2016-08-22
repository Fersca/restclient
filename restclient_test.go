/*
	@author: Fersca
	@date: 11-19-2015
*/

package restclient

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	//Start the testing webserver
	go restAPI()

	//Wait 500ms
	sleep()

	code := m.Run()

	os.Exit(code)
}

///// Test Cases /////

func TestGetDefault(t *testing.T) {

	//Do the GET call
	response, err := Get("http://localhost:8080/testing")

	//Checks if there was an error
	if err != nil {
		t.Fatal("We got an error", err)
	}

	//Checks if the response was 200OK
	if response.Code != 200 {
		t.Fatal("There was not 200OK", "200", response.Code)
	}

	//Checks if the content of the body is as expected
	if response.Body != "{\"id\":\"MLA\"}" {
		t.Fatal("The content was not as expected", "{\"id\":\"MLA\"}", response.Body)
	}

	//Cheks the headers
	if response.Headers == nil {
		t.Fatal("Headers cant be nil")
	}

	//Check the content of one header
	contentType := response.Headers["Content-Type"]
	if contentType[0] != "application/json;charset=UTF-8" {
		t.Fatal("Content-Type:", contentType, "application/json;charset=UTF-8")
	}

	//The end
	fmt.Println("End TestGetDefault")

}

func TestGetDefaultWithHeaders(t *testing.T) {
	//Do the GET call
	response, err := Get("http://localhost:8080/testing", Header{Key: "Test", Value: "Test Header"})

	//Checks if there was an error
	if err != nil {
		t.Fatal("We got an error", err)
	}

	//Checks if the response was 200OK
	if response.Code != 200 {
		t.Fatal("There was not 200OK", "200", response.Code)
	}

	//Checks if the content of the body is as expected
	if response.Body != "{\"id\":\"MLA\"}" {
		t.Fatal("The content was not as expected", "{\"id\":\"MLA\"}", response.Body)
	}

	//Cheks the headers
	if response.Headers == nil {
		t.Fatal("Headers cant be nil")
	}

	//Check the content of one header
	contentType := response.Headers["Test"]
	if contentType[0] != "Test Header" {
		t.Fatal("The content was not as expected", "Test Header", contentType[0])
	}

	//The end
	fmt.Println("End TestGetDefaultWithHeaders")

}

func TestCustomPool(t *testing.T) {
	itemsPool := new(PoolConfig)
	itemsPool.BaseURL = "http://items.mercadolibre.com"

	AddCustomPool("/items/.*", itemsPool)

	usersPool := new(PoolConfig)
	usersPool.BaseURL = "http://users.mercadolibre.com"

	AddCustomPool("/users/.*", usersPool)

	categoriesPool := new(PoolConfig)
	categoriesPool.BaseURL = "http://categories.mercadolibre.com"

	AddCustomPool("/categories/.*", categoriesPool)

	client := getPool("http://items.mercadolibre.com/items/MLA1")

	if client.baseURL != itemsPool.BaseURL {
		t.Fatalf("The content was not as expected, expected: %s, got: %s", itemsPool.BaseURL, client.baseURL)
	}

	client = getPool("http://users.mercadolibre.com/users/1")

	if client.baseURL != usersPool.BaseURL {
		t.Fatalf("The content was not as expected, expected: %s, got: %s", itemsPool.BaseURL, client.baseURL)
	}

	client = getPool("http://categories.mercadolibre.com/categories/MLA3530")

	if client.baseURL != categoriesPool.BaseURL {
		t.Fatalf("The content was not as expected, expected: %s, got: %s", itemsPool.BaseURL, client.baseURL)
	}
}

func TestGetWithCustomPool(t *testing.T) {

	//Create a custom config
	config := new(PoolConfig)
	config.MaxIdleConnsPerHost = 20
	config.Timeout = 0
	config.CacheElements = 0

	//Create a new pool with 5ms of timeout
	AddCustomPool("http://localhost:8080", config)

	//Do the GET call
	response, err := Get("http://localhost:8080/testing")

	//Checks if there was an error
	if err != nil {
		t.Fatal("We got an error", err)
	}

	//Checks if the response was 200OK
	if response.Code != 200 {
		t.Fatal("There was not 200OK", "200", response.Code)
	}

	//Checks if the content of the body is as expected
	if response.Body != "{\"id\":\"MLA\"}" {
		t.Fatal("The content was not as expected", "{\"id\":\"MLA\"}", response.Body)
	}

	//Cheks the headers
	if response.Headers == nil {
		t.Fatal("Headers cant be nil")
	}

	//Check the content of one header
	contentType := response.Headers["Content-Type"]
	if contentType[0] != "application/json;charset=UTF-8" {
		t.Fatal("Content-Type:", contentType, "application/json;charset=UTF-8")
	}

	//The end
	fmt.Println("End TestGetWithCustomPool")

}

func TestGetWithPatternInCustomPool(t *testing.T) {

	//Create a custom config
	config := new(PoolConfig)
	config.BaseURL = "http://localhost:8080"
	config.MaxIdleConnsPerHost = 20
	config.Timeout = 0
	config.CacheElements = 0

	//Create a new pool with 5ms of timeout
	AddCustomPool("/testing.*", config)

	//Do the GET call
	response, err := Get("/testing")

	//Checks if there was an error
	if err != nil {
		t.Fatal("We got an error", err)
	}

	//Checks if the response was 200OK
	if response.Code != 200 {
		t.Fatal("There was not 200OK", "200", response.Code)
	}

	//Checks if the content of the body is as expected
	if response.Body != "{\"id\":\"MLA\"}" {
		t.Fatal("The content was not as expected", "{\"id\":\"MLA\"}", response.Body)
	}

	//Cheks the headers
	if response.Headers == nil {
		t.Fatal("Headers cant be nil")
	}

	//Check the content of one header
	contentType := response.Headers["Content-Type"]
	if contentType[0] != "application/json;charset=UTF-8" {
		t.Fatal("Content-Type:", contentType, "application/json;charset=UTF-8")
	}

	//The end
	fmt.Println("End TestGetWithPatternInCustomPool")

}

func TestGetWithCustomPoolWithTimeout(t *testing.T) {

	//Create a custom config
	config := new(PoolConfig)
	config.MaxIdleConnsPerHost = 20
	config.Timeout = 1

	//Create a new pool with 5ms of timeout
	AddCustomPool("http://localhost:8080", config)

	//Do the GET call
	_, err := Get("http://localhost:8080/testing")

	//Checks if there was an error
	if err == nil {
		t.Fatal("We should had got a timeout", err)
	}

	//Restore the pool
	config.Timeout = 0
	AddCustomPool("http://localhost:8080", config)

	//The end
	fmt.Println("End TestGetWithCustomPoolWithTimeout")

}

func TestGetWithCache(t *testing.T) {

	//Test what happend if there is a Cache-Control header = 10

	//Create a custom config
	config := new(PoolConfig)
	config.CacheElements = 100

	//Create a new pool with cache
	AddCustomPool("http://localhost:8080", config)

	//Do the GET call
	response, _ := Get("http://localhost:8080/cache?seconds=10")

	//Checks if the content of the body is as expected and was not from the cache
	if response.Body != "{\"id\":\"MLA\"}" || response.CachedContent == true {
		t.Fatal("The content was not as expected", "{\"id\":\"MLA\"}", response.Body)
	}

	//Do the GET call
	cachedResponse, _ := Get("http://localhost:8080/cache?seconds=10")

	//Checks if the content of the body is as expected and was from the cache
	if cachedResponse.Body != "{\"id\":\"MLA\"}" || cachedResponse.CachedContent == false {
		fmt.Println("Cached content:", cachedResponse.CachedContent)
		fmt.Println("Code:", cachedResponse.Code)
		t.Fatal("The content was not as expected or was not cached", "{\"id\":\"MLA\"}", cachedResponse.Body)
	}

	//Test what happend if there is not Cache-Control header

	//Reset the cache
	config.Timeout = 0
	config.CacheElements = 100
	AddCustomPool("http://localhost:8080", config)

	//Do the GET call (without cache control)
	response, _ = Get("http://localhost:8080/cache")

	//Checks if the content of the body is as expected and was not from the cache
	if response.Body != "{\"id\":\"MLA\"}" || response.CachedContent == true {
		t.Fatal("The content was not as expected", "{\"id\":\"MLA\"}", response.Body)
	}

	//Do the GET call again and check that the last call was not cached
	cachedResponse, _ = Get("http://localhost:8080/cache")

	//Checks if the content of the body is as expected and was from the cache
	if cachedResponse.Body != "{\"id\":\"MLA\"}" || cachedResponse.CachedContent == true {
		t.Fatal("The content was not as expected", "{\"id\":\"MLA\"}", cachedResponse.Body)
	}

	//Test what happend if the response has a Cache-Control=0

	//Reset the cache
	config.Timeout = 0
	config.CacheElements = 100
	AddCustomPool("http://localhost:8080", config)

	//Do the GET call (without cache control)
	response, _ = Get("http://localhost:8080/cache?seconds=0")

	//Checks if the content of the body is as expected and was not from the cache
	if response.Body != "{\"id\":\"MLA\"}" || response.CachedContent == true {
		t.Fatal("The content was not as expected", "{\"id\":\"MLA\"}", response.Body)
	}

	//Do the GET call again and check that the last call was not cached
	cachedResponse, _ = Get("http://localhost:8080/cache?seconds=0")

	//Checks if the content of the body is as expected and was from the cache
	if cachedResponse.Body != "{\"id\":\"MLA\"}" || cachedResponse.CachedContent == true {
		t.Fatal("The content was not as expected", "{\"id\":\"MLA\"}", cachedResponse.Body)
	}

	//Reset the cache to (not cache)
	config.Timeout = 0
	AddCustomPool("http://localhost:8080", config)

	//The end
	fmt.Println("End TestGetWithCache")

}

func TestGetWithStaleCache(t *testing.T) {

	//Create a custom config
	config := new(PoolConfig)
	config.CacheElements = 100
	config.CacheState = true

	//Create a new pool with cache
	AddCustomPool("http://localhost:8080", config)

	//Do the GET call
	response, _ := Get("http://localhost:8080/cache?seconds=1")

	//Checks if the content of the body is as expected and was not from the cache
	if response.Body != "{\"id\":\"MLA\"}" || response.CachedContent == true {
		t.Fatal("The content was not as expected", "{\"id\":\"MLA\"}", response.Body)
	}

	//Do the GET call
	cachedResponse, _ := Get("http://localhost:8080/cache?seconds=1")

	//Checks if the content of the body is as expected and was from the cache
	if cachedResponse.Body != "{\"id\":\"MLA\"}" || cachedResponse.CachedContent == false || cachedResponse.Staled == true {
		fmt.Println("Cached content:", cachedResponse.CachedContent)
		fmt.Println("Code:", cachedResponse.Code)
		t.Fatal("The content was not as expected or was not cached", "{\"id\":\"MLA\"}", cachedResponse.Body)
	}

	//Sleep for 2 seconds to expire the cachedResponse
	time.Sleep(1100 * time.Millisecond)

	//Do the GET call and we hope to get the stale response (adding the header for 500 error response)
	staledResponse, _ := Get("http://localhost:8080/cache?seconds=1", Header{Key: "Error", Value: "500 error"})

	//Checks if the content of the body is as expected and was from the cache
	if staledResponse.Staled == false {
		fmt.Println("Cached content:", staledResponse.CachedContent)
		fmt.Println("Code:", staledResponse.Code)
		fmt.Println("Staled:", staledResponse.Staled)
		t.Fatal("The content was not as expected or was not cached", "{\"id\":\"MLA\"}", staledResponse.Body)
	}

	//Reset the cache to (not cache)
	config.Timeout = 0
	AddCustomPool("http://localhost:8080", config)

	//The end
	fmt.Println("End TestGetWithStaleCache")

}

func TestPostDefault(t *testing.T) {

	//Do the POST call
	response, err := Post("http://localhost:8080/testing", "{\"id\":\"MLA\"}")

	//Checks if there was an error
	if err != nil {
		t.Fatal("We got an error", err)
	}

	//Checks if the response was 201OK
	if response.Code != 201 {
		t.Fatal("There was not 201OK", "201", response.Code)
	}

	//Checks if the content of the body is as expected
	if response.Body != "echo --> {\"id\":\"MLA\"}" {
		t.Fatal("The content was not as expected", "echo --> {\"id\":\"MLA\"}", response.Body)
	}

	//Cheks the headers
	if response.Headers == nil {
		t.Fatal("Headers cant be nil")
	}

	//Check the content of one header
	contentType := response.Headers["Content-Type"]
	if contentType[0] != "application/json;charset=UTF-8" {
		t.Fatal("Content-Type:", contentType, "application/json;charset=UTF-8")
	}

	//The end
	fmt.Println("End TestPostDefault")

}

func TestPutDefault(t *testing.T) {

	//Do the PUT call
	response, err := Put("http://localhost:8080/testing", "{\"id\":\"MLA\"}")

	//Checks if there was an error
	if err != nil {
		t.Fatal("We got an error", err)
	}

	//Checks if the response was 201OK
	if response.Code != 200 {
		t.Fatal("There was not 201OK", "201", response.Code)
	}

	//Checks if the content of the body is as expected
	if response.Body != "echoPut --> {\"id\":\"MLA\"}" {
		t.Fatal("The content was not as expected", "echo --> {\"id\":\"MLA\"}", response.Body)
	}

	//Cheks the headers
	if response.Headers == nil {
		t.Fatal("Headers cant be nil")
	}

	//Check the content of one header
	contentType := response.Headers["Content-Type"]
	if contentType[0] != "application/json;charset=UTF-8" {
		t.Fatal("Content-Type:", contentType, "application/json;charset=UTF-8")
	}

	//The end
	fmt.Println("End TestPutDefault")

}

func TestDeleteDefault(t *testing.T) {

	//Do the Delete call
	response, err := Delete("http://localhost:8080/testing")

	//Checks if there was an error
	if err != nil {
		t.Fatal("We got an error", err)
	}

	//Checks if the response was 201OK
	if response.Code != 200 {
		t.Fatal("There was not 200OK", "200", response.Code)
	}

	//Checks if the content of the body is as expected
	if response.Body != "echoDelete --> OK" {
		t.Fatal("The content was not as expected", "echoDelete --> OK", response.Body)
	}

	//Cheks the headers
	if response.Headers == nil {
		t.Fatal("Headers cant be nil")
	}

	//The end
	fmt.Println("End TestDeleteDefault")

}

func TestHeadDefault(t *testing.T) {

	//Do the Head call
	response, err := Head("http://localhost:8080/testing")

	//Checks if there was an error
	if err != nil {
		t.Fatal("We got an error", err)
	}

	//Checks if the response was 201OK
	if response.Code != 200 {
		t.Fatal("There was not 200OK", "200", response.Code)
	}

	//The end
	fmt.Println("End TestHeadDefault")

}

func TestMock(t *testing.T) {

	//Create the mock response
	mockResp := new(Response)
	mockResp.Body = "{\"id\":\"FER\"}"
	mockResp.CachedContent = false
	mockResp.Code = 200
	mockResp.Headers = nil
	mockResp.Staled = false

	AddMock("http://fer.com", "GET", "", *mockResp)

	//Do the Get call
	response, err := Get("http://fer.com")

	//Checks if there was an error
	if err != nil {
		t.Fatal("We got an error", err)
	}

	//Checks if the response was 20OK
	if response.Code != 200 {
		t.Fatal("There was not 200OK", "200", response.Code)
	}

	//Checks if the content of the body is as expected and was not from the cache
	if response.Body != "{\"id\":\"FER\"}" || response.CachedContent == true {
		t.Fatal("The content was not as expected", "{\"id\":\"MLA\"}", response.Body)
	}

	//The end
	fmt.Println("End TestMock")

}

func TestMockWithHeaders(t *testing.T) {

	headers := []Header{Header{Key: "Accept", Value: "application/json"}, Header{Key: "Encode", Value: "true"}}

	//Create the mock response
	mockResp := new(Response)
	mockResp.Body = "{\"id\":\"Vale\"}"
	mockResp.CachedContent = false
	mockResp.Code = 200
	mockResp.Headers = nil
	mockResp.Staled = false

	AddMock("http://fer.com", "GET", "", *mockResp, headers...)

	//Do the Get call
	response, err := Get("http://fer.com", headers...)

	//Checks if there was an error
	if err != nil {
		t.Fatal("We got an error", err)
	}

	//Checks if the response was 201OK
	if response.Code != 200 {
		t.Fatal("There was not 200OK", "200", response.Code)
	}

	//Checks if the content of the body is as expected and was not from the cache
	if response.Body != "{\"id\":\"Vale\"}" || response.CachedContent == true {
		t.Fatal("The content was not as expected", "{\"id\":\"Vale\"}", response.Body)
	}

	//The end
	fmt.Println("End TestMockWithHeaders")

}

func TestSeveralGETMocks(t *testing.T) {

	mocks := make(map[string]Response)

	//Create the mock response
	mocks["http://fer2.com"] = Response{Body: "{\"id\":\"Fer\"}", Code: 200}
	mocks["http://vale2.com"] = Response{Body: "{\"id\":\"Vale\"}", Code: 201}
	mocks["http://fer/.*/pipi"] = Response{Body: "{\"id\":\"Artu\"}", Code: 200}

	//Add the mocks to the API
	AddMocks(mocks)

	//Set the regex urls
	URLasRegexp("http://fer/.*/pipi")

	//Do the Get call
	response, err := Get("http://fer2.com")

	//Checks if there was an error
	if err != nil {
		t.Fatal("We got an error", err)
	}

	//Checks if the response was 200OK
	if response.Code != 200 {
		t.Fatal("There was not 200OK", "200", response.Code)
	}

	//Checks if the content of the body is as expected and was not from the cache
	if response.Body != "{\"id\":\"Fer\"}" || response.CachedContent == true {
		t.Fatal("The content was not as expected", "{\"id\":\"Fer\"}", response.Body)
	}

	//Do the Get call
	response, err = Get("http://vale2.com")

	//Checks if there was an error
	if err != nil {
		t.Fatal("We got an error", err)
	}

	//Checks if the response was 201
	if response.Code != 201 {
		t.Fatal("There was not 201OK", "201", response.Code)
	}

	//Checks if the content of the body is as expected and was not from the cache
	if response.Body != "{\"id\":\"Vale\"}" || response.CachedContent == true {
		t.Fatal("The content was not as expected", "{\"id\":\"Vale\"}", response.Body)
	}

	//Do the Get call
	response, err = Get("http://fer/jamon/pipi")

	//Checks if there was an error
	if err != nil {
		t.Fatal("We got an error", err)
	}

	//Checks if the response was 200
	if response.Code != 200 {
		t.Fatal("There was not 200OK", "200", response.Code)
	}

	//Checks if the content of the body is as expected and was not from the cache
	if response.Body != "{\"id\":\"Artu\"}" || response.CachedContent == true {
		t.Fatal("The content was not as expected", "{\"id\":\"Artu\"}", response.Body)
	}

	//The end
	fmt.Println("End TestSeveralGETMocks")

}

///// Utils /////

//Sleep for 100ms
func sleep() {
	time.Sleep(100 * time.Millisecond)
}

//Init for testing
func restAPI() {
	//Create the webserver
	http.Handle("/testing", http.HandlerFunc(processRequestDefault))
	http.Handle("/cache", http.HandlerFunc(processRequestCache))

	err := http.ListenAndServe("0.0.0.0:8080", nil)
	if err != nil {
		fmt.Print("Error", err)
	}
}

//processRequest process the request for the mock webserver
func processRequestCache(w http.ResponseWriter, req *http.Request) {

	//Get the headers map
	headerMap := w.Header()

	parameter := req.URL.Query()["seconds"]
	if parameter != nil {
		seconds := req.URL.Query()["seconds"][0]

		if seconds != "" {
			headerMap.Add("Cache-Control", "max-age="+seconds)
		}
	}

	//Copy headers sent to the response
	if req.Header["Error"] != nil {
		w.WriteHeader(500)
		w.Write([]byte("{\"error\":\"Error getting resource\"}"))
	} else {
		w.WriteHeader(200)
		w.Write([]byte("{\"id\":\"MLA\"}"))
	}

	return

}

//processRequest process the request for the mock webserver
func processRequestDefault(w http.ResponseWriter, req *http.Request) {

	//Get the headers map
	headerMap := w.Header()

	//Returns the payload in the response (echo)
	headerMap.Add("Content-Type", "application/json;charset=UTF-8")

	//Copy headers sent to the response
	if req.Header["Test"] != nil {
		headerMap.Add("Test", req.Header["Test"][0])
	}

	//Performs action based on the request Method
	switch req.Method {

	case http.MethodGet:

		//Wait 100ms
		sleep()

		//return the example json
		w.WriteHeader(200)
		w.Write([]byte("{\"id\":\"MLA\"}"))
		return

	case http.MethodHead:

		//return the example json
		w.WriteHeader(200)
		return

	case http.MethodPut:

		//Create the array to hold the body
		p := make([]byte, req.ContentLength)

		//Reads the body content
		req.Body.Read(p)

		w.WriteHeader(200)
		w.Write([]byte("echoPut --> " + string(p)))

	case http.MethodDelete:
		w.WriteHeader(200)
		w.Write([]byte("echoDelete --> OK"))

	case http.MethodPost:

		//Create the array to hold the body
		p := make([]byte, req.ContentLength)

		//Reads the body content
		req.Body.Read(p)

		w.WriteHeader(201)
		w.Write([]byte("echo --> " + string(p)))

	default:
		//Method Not Allowed
		w.WriteHeader(405)
	}

}
