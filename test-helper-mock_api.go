//helpers
package main

import (
	"io/ioutil"
	"github.com/h2non/gock"
)

func MockApiCall (request string, http_status int, fixture string) {
	// load fixture
	file, _ := ioutil.ReadFile(fixture)
	listStacksFixture := string(file)	

	gock.Off()
	gock.New("https://app.cloud66.com/api/3").
		Get(request).
		Reply(http_status).
		BodyString(listStacksFixture)
 }