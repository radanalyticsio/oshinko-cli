package unittest

import (
	"regexp"

	"gopkg.in/check.v1"

	"github.com/radanalyticsio/oshinko-cli/rest/helpers/uuid"
)

func (s *OshinkoUnitTestSuite) TestUuid(c *check.C) {
	var validVer4UUID = regexp.MustCompile(`[a-z0-9]{8}-[a-z0-9]{4}-4[a-z0-9]{3}-[ab89][a-z0-9]{3}-[a-z0-9]{12}`)

	uuid, err := uuid.Uuid()

	c.Assert(nil, check.Equals, err)
	c.Assert(true, check.Equals, validVer4UUID.MatchString(uuid))
}
