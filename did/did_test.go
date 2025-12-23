package did

import (
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGenerateDistributedID(t *testing.T) {
	Convey("test generate distributed id\n", t, func() {

		Convey("ipv4: 1.2.3.4\n", func() {
			os.Setenv("POD_IP", "1.2.3.4")
			id, err := GenerateDistributedID()

			So(id, ShouldNotEqual, 0)
			So(err, ShouldBeNil)
		})

		Convey("ipv6: fc00:b36f:c1c3:2000:f821:712f:c2f8:6026\n", func() {
			os.Setenv("POD_IP", "fc00:b36f:c1c3:2000:f821:712f:c2f8:6026")
			id, err := GenerateDistributedID()

			So(id, ShouldNotEqual, 0)
			So(err, ShouldBeNil)
		})

		Convey("ipv6: fc99:3504::a04:68cc\n", func() {
			os.Setenv("POD_IP", "fc99:3504::a04:68cc")
			id, err := GenerateDistributedID()

			So(id, ShouldNotEqual, 0)
			So(err, ShouldBeNil)
		})
	})
}
