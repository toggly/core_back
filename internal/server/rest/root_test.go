package rest_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Toggly/core/internal/pkg/api"
	"github.com/Toggly/core/internal/pkg/storage/mongo"
	"github.com/Toggly/core/internal/server/rest"
	"github.com/globalsign/mgo"
	asserts "github.com/stretchr/testify/assert"
)

const MongoTestUrl = "mongodb://localhost:27017/toggly_test_rest"
const TestAuthToken = "TestToken"
const ow = "test_owner"

type TestCase struct {
	name         string
	method       string
	requestId    string
	cType        string
	path         string
	status       int
	before       func(rs *httptest.Server)
	after        func(rs *httptest.Server)
	validator    func(body []byte)
	patchRequest func(req *http.Request)
	skip         bool
	body         interface{}
}

func GetRouter() *rest.APIRouter {
	dataStorage, _ := mongo.NewMongoStorage(MongoTestUrl)
	engine := &api.Engine{Storage: &dataStorage}
	return &rest.APIRouter{
		Version:   "test",
		Cache:     nil,
		Engine:    engine,
		BasePath:  "/api",
		IsDebug:   false,
		AuthToken: TestAuthToken,
	}
}

func DropDB() {
	session, _ := mgo.Dial(MongoTestUrl)
	session.DB("").DropDatabase()
}

func BeforeTest() {
	DropDB()
}

func AfterTest() {
	DropDB()
}

func bodyJSON(body []byte) (map[string]interface{}, error) {
	var dat map[string]interface{}
	if err := json.Unmarshal(body, &dat); err != nil {
		return nil, err
	}
	return dat, nil
}

func parseBodyTo(body []byte, obj interface{}) error {
	if err := json.Unmarshal(body, obj); err != nil {
		return err
	}
	return nil
}

func bodyText(body []byte) string {
	return string(bytes.TrimSpace(body))
}

func getBody(r *http.Response) []byte {
	defer r.Body.Close()
	byt, _ := ioutil.ReadAll(r.Body)
	return byt
}

func runTestCase(t *testing.T, rs *httptest.Server, tc TestCase) {
	assert := asserts.New(t)
	t.Run(tc.name, func(t *testing.T) {
		if tc.skip {
			t.Skip()
		}
		if tc.before != nil {
			tc.before(rs)
		}
		var body []byte
		var err error
		if tc.body != nil {
			body, err = json.Marshal(tc.body)
			assert.Nil(err)
		}
		req, err := http.NewRequest(tc.method, rs.URL+tc.path, bytes.NewBuffer(body))
		assert.Nil(err)
		req.Header = http.Header{
			rest.XTogglyAuth:    []string{TestAuthToken},
			rest.XTogglyOwnerID: []string{ow},
		}
		if tc.requestId != "" {
			req.Header[rest.XTogglyRequestID] = []string{tc.requestId}
		}
		if tc.patchRequest != nil {
			tc.patchRequest(req)
		}
		r, err := rs.Client().Do(req)
		fmt.Printf("\nRESP: %v\n\n", r)
		responceBody := getBody(r)
		fmt.Printf("BODY: %s\n", responceBody)
		assert.Nil(err)
		assert.NotNil(r)
		assert.Equal(tc.status, r.StatusCode)
		assert.Equal("Toggly", r.Header[rest.XServiceName][0])
		assert.Equal("test", r.Header[rest.XServiceVersion][0])
		if tc.requestId != "" {
			assert.Equal(tc.requestId, r.Header[rest.XTogglyRequestID][0])
		} else {
			assert.Regexp("^req-\\d*$", r.Header[rest.XTogglyRequestID][0])
		}
		if tc.cType == "" {
			tc.cType = "application/json"
		}
		assert.Contains(r.Header[http.CanonicalHeaderKey("Content-Type")][0], tc.cType)
		if tc.validator != nil {
			tc.validator(responceBody)
		}
		if tc.after != nil {
			tc.after(rs)
		}
	})
}

func TestRestRequestHeaders(t *testing.T) {
	assert := asserts.New(t)
	BeforeTest()

	tt := []TestCase{
		{
			name:   "no auth",
			method: http.MethodGet,
			path:   "/api/v1",
			status: http.StatusUnauthorized,
			cType:  "text/plain",
			patchRequest: func(req *http.Request) {
				req.Header[rest.XTogglyAuth] = nil
			},
		},
		{
			name:   "wrong token",
			method: http.MethodGet,
			path:   "/api/v1",
			status: http.StatusUnauthorized,
			cType:  "text/plain",
			patchRequest: func(req *http.Request) {
				req.Header[rest.XTogglyAuth] = []string{"wrong_token"}
			},
		},
		{
			name:   "authorized but owner not found",
			method: http.MethodGet,
			path:   "/api/v1",
			status: http.StatusNotFound,
			patchRequest: func(req *http.Request) {
				req.Header[rest.XTogglyOwnerID] = nil
			},
			validator: func(body []byte) {
				b, err := bodyJSON(body)
				assert.Nil(err)
				assert.Equal("Owner not found", b["error"])
			},
		},
		{
			name:   "authorized",
			method: http.MethodGet,
			path:   "/api/v1",
			status: http.StatusNotFound,
			cType:  "text/plain",
			validator: func(body []byte) {
				assert.Equal("404 page not found", bodyText(body))
			},
		},
	}

	rs := httptest.NewServer(GetRouter().Router())
	defer rs.Close()

	for _, tc := range tt {
		runTestCase(t, rs, tc)
	}

	AfterTest()
}