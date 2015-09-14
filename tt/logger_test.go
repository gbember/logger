// logger_test.go
package tt

import (
	"log"
	"testing"
	"time"

	"github.com/gbember/logger"
)

func TestLogger(t *testing.T) {
	err := logger.StartLog("./", logger.DEBUG)
	if err != nil {
		t.Fatal(err)
	}
	logger.Debug("======测试%s  %d", "sgas", 1)
	log.Println("jsagisajgiswgoijweghwiohj")
	time.Sleep(time.Second * 1)
}
