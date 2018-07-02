package fake

// Client manages communication with the Kong API.
// New client objects should be created using the NewClient function.
// The BaseURL field must be defined and pointed at an instance of the
// Kong Admin API.
//
// Kong resources can be access using the Service objects.
// client.Apis.Get("id") -> GET /apis/id
// client.Consumers.Patch(consumer) -> PATCH /consumers/id
type Client struct {
	common service

	//Route   *RouteService
	Service *ServiceService
}

// Each service representing a Kong resource type will be of this type
type service struct {
	client        *Client
	ExpectedOuput interface{}
}

// NewClient creates a new kong.Client object.
// This should be the primary way a kong.Client object is constructed.
//
// If an httpClient object is specified it will be used instead of the
// default http.DefaultClient.
//
// baseURLStr should point to an instance a Kong Admin API and must
// contain the trailing slash. i.e. http://kong:8001/
func NewClient() (*Client, error) {
	c := &Client{}
	c.common.client = c
	//c.Route = (*RouteService)(&c.common)
	c.Service = (*ServiceService)(&c.common)

	return c, nil
}
