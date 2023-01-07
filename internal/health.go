package internal

import (
	"fmt"
	gosundheit "github.com/AppsFlyer/go-sundheit"
	"github.com/AppsFlyer/go-sundheit/checks"
	healthHttp "github.com/AppsFlyer/go-sundheit/http"
	netHttp "net/http"
	"strings"
	"time"
)

// RegisterHealthChecks initializes health checks for the Exporter and returns an http handler func
func RegisterHealthChecks(mongoUrl string) (netHttp.HandlerFunc, error) {

	mongoHttp := strings.Replace(mongoUrl, "mongodb://", "http://", 1)
	httpCheckConf := checks.HTTPCheckConfig{
		CheckName: "mongo.url.check",
		Timeout:   1 * time.Second,
		URL:       mongoHttp,
	}

	httpCheck, err := checks.NewHTTPCheck(httpCheckConf)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	h := gosundheit.New()
	err = h.RegisterCheck(
		httpCheck,
		gosundheit.InitialDelay(time.Second),
		gosundheit.ExecutionPeriod(10*time.Second),
	)

	if err != nil {
		fmt.Println("Failed to register check: ", err)
		return nil, err
	}

	return healthHttp.HandleHealthJSON(h), nil
}
