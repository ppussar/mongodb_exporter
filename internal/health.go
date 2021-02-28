package internal

import (
	"fmt"
	"github.com/AppsFlyer/go-sundheit"
	"github.com/AppsFlyer/go-sundheit/checks"
	healthHttp "github.com/AppsFlyer/go-sundheit/http"
	netHttp "net/http"
	"strings"
	"time"
)

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
	httpCheck = checks.Must(checks.NewHTTPCheck(httpCheckConf))

	h := health.New()
	err = h.RegisterCheck(&health.Config{
		Check:           httpCheck,
		ExecutionPeriod: 10 * time.Second, // the check will be executed every 10 sec
	})

	if err != nil {
		fmt.Println("Failed to register check: ", err)
		return nil, err
	}

	return healthHttp.HandleHealthJSON(h), nil
}
