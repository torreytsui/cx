//helpers
package main

import (
	"github.com/h2non/gock"
	"io/ioutil"
)

func MockApiGetCall(request string, http_status int, fixture string) {
	// load fixture
	file, _ := ioutil.ReadFile(fixture)
	listStacksFixture := string(file)

	gock.New("https://app.cloud66.com/api/3").
		Get(request).
		Reply(http_status).
		BodyString(listStacksFixture)
}

func MockApiPutCall(request string, http_status int, fixture string) {
	// load fixture
	file, _ := ioutil.ReadFile(fixture)
	listStacksFixture := string(file)

	gock.New("https://app.cloud66.com/api/3").
		Put(request).
		Reply(http_status).
		BodyString(listStacksFixture)
}
