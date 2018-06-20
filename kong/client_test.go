package kong

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
)

var (
	// HTTP mux used with test server
	mux *http.ServeMux

	// Kong client being tested
	client *Client

	// Test server used to stub Kong resources
	server *httptest.Server
)

// stubSetup creates a test HTTP server and a kong.Client that is
// configured to talk to the test server.
//
// Tests should register handlers which provide stub responses for
// the API method being tested.
func stubSetup() {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	client, _ = NewClient(nil, server.URL)
}

func stubTeardown() {
	server.Close()
}

// testMethod is used in stub http.HandleFunc's to check if
// the appropriate HTTP Method was used by the client.
func testMethod(t *testing.T, r *http.Request, want string) {
	if got := r.Method; got != want {
		t.Errorf("Request method: %v, want %v", got, want)
	}
}

type values map[string]string

// testFormValues is used in stub http.HandleFunc's to check
// if the correct URI parameters were provided
func testFormValues(t *testing.T, r *http.Request, values values) {
	want := url.Values{}
	for k, v := range values {
		want.Add(k, v)
	}

	r.ParseForm()
	if got := r.Form; !reflect.DeepEqual(got, want) {
		t.Errorf("Request parameters: %v, want %v", got, want)
	}
}

// testFormValues is used in stub http.HandleFunc's to check
// if the correct Header was provided.
func testHeader(t *testing.T, r *http.Request, header string, want string) {
	if got := r.Header.Get(header); got != want {
		t.Errorf("Header.Get(%q) returned %q, want %q", header, got, want)
	}
}

func testURLParseError(t *testing.T, err error) {
	if err == nil {
		t.Error("Expected error to be returned")
	}
	if err, ok := err.(*url.Error); !ok || err.Op != "parse" {
		t.Errorf("Expected URL parse error, got %+v", err)
	}
}

// testBody is used in stub http.HandleFunc's to check
// if the correct request body was provided.
func testBody(t *testing.T, r *http.Request, want string) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		t.Errorf("Error reading request body: %v", err)
	}
	if got := string(b); got != want {
		t.Errorf("request Body is %s, want %s", got, want)
	}
}

// Helper function to test that a value is marshaled to JSON as expected.
func testJSONMarshal(t *testing.T, v interface{}, want string) {
	j, err := json.Marshal(v)
	if err != nil {
		t.Errorf("Unable to marshal JSON for %v", v)
	}

	w := new(bytes.Buffer)
	err = json.Compact(w, []byte(want))
	if err != nil {
		t.Errorf("String is` not valid json: %s", want)
	}

	if w.String() != string(j) {
		t.Errorf("json.Marshal(%q) returned %s, want %s", v, j, w)
	}

	// now go the other direction and make sure things unmarshal as expected
	u := reflect.ValueOf(v).Interface()
	if err := json.Unmarshal([]byte(want), u); err != nil {
		t.Errorf("Unable to unmarshal JSON for %v", want)
	}

	if !reflect.DeepEqual(v, u) {
		t.Errorf("json.Unmarshal(%q) returned %s, want %s", want, u, v)
	}
}

func TestNewClient(t *testing.T) {
	c, _ := NewClient(nil, "http://test:8001")

	if got, want := c.BaseURL.String(), "http://test:8001"; got != want {
		t.Errorf("NewClient BaseURL is %v, want %v", got, want)
	}

}

func TestNewClient_badUrlStr(t *testing.T) {
	_, err := NewClient(nil, "%")
	if err == nil {
		t.Error("Expected error to be returned.")
	}
}

const defaultBaseURL = "http://test:8001/"

// func TestNewRequest(t *testing.T) {
// 	c, _ := NewClient(nil, defaultBaseURL)

// 	inURL, outURL := "foo", defaultBaseURL+"foo"
// 	//inBody, outBody := &Api{Name: "n"}, `{"name":"n"}`+"\n"
// 	req, _ := c.NewRequest("GET", inURL, inBody)

// 	// Test that relative URL was expanded
// 	if got, want := req.URL.String(), outURL; got != want {
// 		t.Errorf("NewRequest(%q) URL is %v, want %v", inURL, got, want)
// 	}

// 	// Test that body was JSON encoded
// 	body, _ := ioutil.ReadAll(req.Body)
// 	if got, want := string(body), outBody; got != want {
// 		t.Errorf("NewRequest(%v) Body is %v, want %v", inBody, got, want)
// 	}
// }

func TestNewRequest_invalidJSON(t *testing.T) {
	c, _ := NewClient(nil, defaultBaseURL)

	type T struct {
		A map[interface{}]interface{}
	}
	_, err := c.NewRequest("GET", "/", &T{})

	if err == nil {
		t.Error("Expected error to be returned.")
	}
	if err, ok := err.(*json.UnsupportedTypeError); !ok {
		t.Errorf("Expected a JSON error; got %#v.", err)
	}
}

func TestNewRequest_badURL(t *testing.T) {
	c, _ := NewClient(nil, defaultBaseURL)
	_, err := c.NewRequest("GET", ":", nil)
	testURLParseError(t, err)
}

// If a nil body is passed to kong.NewRequest, make sure that nil is also
// passed to http.NewRequest.  In most cases, passing an io.Reader that returns
// no content is fine, since there is no difference between an HTTP request
// body that is an empty string versus one that is not set at all.  However in
// certain cases, intermediate systems may treat these differently resulting in
// subtle errors.
func TestNewRequest_emptyBody(t *testing.T) {
	c, _ := NewClient(nil, defaultBaseURL)
	req, err := c.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("NewRequest returned unexpected error: %v", err)
	}
	if req.Body != nil {
		t.Fatal("constructed request contains a non-nil Body")
	}
}

func TestDo(t *testing.T) {
	stubSetup()
	defer stubTeardown()

	type foo struct {
		A string
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if m := "GET"; m != r.Method {
			t.Errorf("Request method = %v, want %v", r.Method, m)
		}
		fmt.Fprint(w, `{"A":"a"}`)
	})

	req, _ := client.NewRequest("GET", "/", nil)
	body := new(foo)
	client.Do(req, body)

	want := &foo{"a"}
	if !reflect.DeepEqual(body, want) {
		t.Errorf("Response body = %v, want %v", body, want)
	}
}

func TestDo_httpError(t *testing.T) {
	stubSetup()
	defer stubTeardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Bad Request", 400)
	})

	req, _ := client.NewRequest("GET", "/", nil)
	_, err := client.Do(req, nil)

	if err == nil {
		t.Error("Expected HTTP 400 error.")
	}
}

func TestDo_noContent(t *testing.T) {
	stubSetup()
	defer stubTeardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	var body json.RawMessage

	req, _ := client.NewRequest("GET", "/", nil)
	_, err := client.Do(req, &body)
	if err != nil {
		t.Fatalf("Do returned unexpected error: %v", err)
	}
}

func TestErrorResponse_Error(t *testing.T) {
	url, _ := url.Parse("http://test")
	e := &ErrorResponse{
		Response: &http.Response{
			StatusCode: 200,
			Request: &http.Request{
				Method: "m",
				URL:    url,
			},
		},
		KongMessage: "m",
		KongError:   "e",
	}
	want := "m http://test: 200 m e"
	if e.Error() != want {
		t.Fatalf("ErrorResponse.Error() did not return the correct string. got %v, want %v", e.Error(), want)
	}
}

func TestConflictError_Error(t *testing.T) {
	url, _ := url.Parse("http://test")
	e := &ConflictError{
		Response: &http.Response{
			StatusCode: 200,
			Request: &http.Request{
				Method: "m",
				URL:    url,
			},
		},
		KongMessage: "m",
		KongError:   "e",
	}
	want := "m http://test: 200 m e"
	if e.Error() != want {
		t.Fatalf("ConflictError.Error() did not return the correct string. got %v, want %v", e.Error(), want)
	}
}

func TestNotFoundError_Error(t *testing.T) {
	url, _ := url.Parse("http://test")
	e := &NotFoundError{
		Response: &http.Response{
			StatusCode: 200,
			Request: &http.Request{
				Method: "m",
				URL:    url,
			},
		},
		KongMessage: "m",
		KongError:   "e",
	}
	want := "m http://test: 200 m e"
	if e.Error() != want {
		t.Fatalf("ConflictError.Error() did not return the correct string. got %v, want %v", e.Error(), want)
	}
}

func TestCheckResponse(t *testing.T) {
	r := &http.Response{
		StatusCode: 200,
	}
	req, _ := http.NewRequest("GET", "/url", nil)
	err := CheckResponse(req, r)
	if err != nil {
		t.Fatalf("CheckResponse returned unexpected error: %v", err)
	}
}

func TestCheckResponse_badStatusCode(t *testing.T) {
	r := &http.Response{
		StatusCode: 400,
		Body:       ioutil.NopCloser(bytes.NewBufferString(`{"error": "e"}`)),
	}
	req, _ := http.NewRequest("GET", "/url", nil)
	err := CheckResponse(req, r)
	_, ok := err.(*ErrorResponse)
	if !ok {
		t.Fatal("CheckResponse returned the incorrect error type")
	}
}

func TestCheckResponse_notFoundStatusCode(t *testing.T) {
	r := &http.Response{
		StatusCode: 404,
		Body:       ioutil.NopCloser(bytes.NewBufferString(`{"error": "e"}`)),
	}
	req, _ := http.NewRequest("GET", "/url", nil)
	err := CheckResponse(req, r)
	_, ok := err.(*NotFoundError)
	if !ok {
		t.Fatal("CheckResponse returned the incorrect error type")
	}
}

func TestCheckResponse_conflictStatusCode(t *testing.T) {
	r := &http.Response{
		StatusCode: 409,
		Body:       ioutil.NopCloser(bytes.NewBufferString(`{"error": "e"}`)),
	}
	req, _ := http.NewRequest("GET", "/url", nil)
	err := CheckResponse(req, r)
	_, ok := err.(*ConflictError)
	if !ok {
		t.Fatal("CheckResponse returned the incorrect error type")
	}
}

// TODO
// func TestCheckResponse(t *testing.T) {
// func TestCheckResponse_noBody(t *testing.T) {
// func TestErrorResponse_Error(t *testing.T) {
// func TestError_Error(t *testing.T) {
