package api

import "github.com/go-kit/kit/endpoint"

// compile time assertions for our response types implementing endpoint.Failer.
var (
	_ endpoint.Failer = PingResponse{}
	_ endpoint.Failer = SignupResponse{}
	_ endpoint.Failer = LoginResponse{}
	_ endpoint.Failer = CreateListResponse{}
)

// Failed implements endpoint.Failer.
func (r PingResponse) Failed() error { return r.Err }

// Failed implements endpoint.Failer.
func (r SignupResponse) Failed() error { return r.Err }

// Failed implements endpoint.Failer.
func (r LoginResponse) Failed() error { return r.Err }

// Failed implements endpoint.Failer.
func (r LogoutResponse) Failed() error { return r.Err }

// Failed implements endpoint.Failer.
func (r CreateListResponse) Failed() error { return r.Err }

// Failed implements endpoint.Failer.
func (r GetListsResponse) Failed() error { return r.Err }

// Failed implements endpoint.Failer.
func (r CreateItemResponse) Failed() error { return r.Err }

// Failed implements endpoint.Failer.
func (r GetListItemsResponse) Failed() error { return r.Err }

// Failed implements endpoint.Failer.
func (r BuyItemResponse) Failed() error { return r.Err }

// Failed implements endpoint.Failer.
func (r ShareListResponse) Failed() error { return r.Err }

// Failed implements endpoint.Failer.
func (r GetAllCategoriesResponse) Failed() error { return r.Err }

// Failed implements endpoint.Failer.
func (r DeleteListResponse) Failed() error { return r.Err }

// Failed implements endpoint.Failer.
func (r DeleteItemResponse) Failed() error { return r.Err }
