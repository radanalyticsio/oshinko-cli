package unittest

import (
	"gopkg.in/check.v1"

	"github.com/redhatanalytics/oshinko-rest/helpers/errors"
)

func (s *OshinkoUnitTestSuite) TestNewSingleErrorReponse(c *check.C) {
	expectedStatus := int32(123)
	expectedTitle := "A good test title"
	expectedDetails := "These are the details of the test"
	observedResponse := errors.NewSingleErrorResponse(
		expectedStatus, expectedTitle, expectedDetails)
	c.Assert(len(observedResponse.Errors), check.Equals, 1)
	c.Assert(*observedResponse.Errors[0].Status, check.Equals, expectedStatus)
	c.Assert(*observedResponse.Errors[0].Title, check.Equals, expectedTitle)
	c.Assert(*observedResponse.Errors[0].Details, check.Equals, expectedDetails)
}
