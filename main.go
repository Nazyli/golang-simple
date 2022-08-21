package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/hellofresh/health-go/v4"
	healthHttp "github.com/hellofresh/health-go/v4/checks/http"
	healthMySql "github.com/hellofresh/health-go/v4/checks/mysql"
	healthPg "github.com/hellofresh/health-go/v4/checks/postgres"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

func init() {
	log.SetPrefix("[API Golang : ] ")
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	_ = godotenv.Load()

	e := echo.New()

	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	sample := e.Group("/status")
	sample.GET("/", Get)
	sample.GET("/health-check", echo.WrapHandler(HealthCheck()))
	sample.GET("/ip", ReadUserIP)

	port, ok := os.LookupEnv("PORT")
	if !ok {
		panic("missing PORT environment")
	}
	listenerPort := fmt.Sprintf(":%s", port)
	e.Logger.Fatal(e.Start(listenerPort))
}

func Get(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": "Server is up and running base api",
		"time": time.Now(),
		//"unix": time.Now().UTC().UnixNano(),
	})
}

func HealthCheck() http.Handler {
	h, _ := health.New()
	h.Register(health.Config{
		Name:      "some-custom-check-fail",
		Timeout:   time.Second * 5,
		SkipOnErr: true,
		Check:     func(context.Context) error { return errors.New("failed during custom health check") },
	})
	h.Register(health.Config{
		Name:  "some-custom-check-success",
		Check: func(context.Context) error { return nil },
	})
	h.Register(health.Config{
		Name:      "http-check",
		Timeout:   time.Second * 5,
		SkipOnErr: true,
		Check: healthHttp.New(healthHttp.Config{
			URL: `http://example.com`,
		}),
	})
	h.Register(health.Config{
		Name:      "postgres-check",
		Timeout:   time.Second * 5,
		SkipOnErr: true,
		Check: healthPg.New(healthPg.Config{
			DSN: `postgres://test:test@0.0.0.0:32783/test?sslmode=disable`,
		}),
	})

	h.Register(health.Config{
		Name:      "mysql-check",
		Timeout:   time.Second * 5,
		SkipOnErr: true,
		Check: healthMySql.New(healthMySql.Config{
			DSN: `test:test@tcp(0.0.0.0:32778)/test?charset=utf8`,
		}),
	})

	h.Register(health.Config{
		Name:      "rabbit-aliveness-check",
		Timeout:   time.Second * 5,
		SkipOnErr: true,
		Check: healthHttp.New(healthHttp.Config{
			URL: `http://guest:guest@0.0.0.0:32780/api/aliveness-test/%2f`,
		}),
	})
	return h.Handler()
}

func ReadUserIP(c echo.Context) error {
	IPAddress := c.Request().Header.Get("X-Real-Ip")
	if IPAddress == "" {
		IPAddress = c.Request().Header.Get("X-Forwarded-For")
	}
	if IPAddress == "" {
		IPAddress = c.Request().RemoteAddr
	}
	ipServer, err := externalIP()
	if err != nil {
		ipServer = err.Error()
	}
	hostname, err := os.Hostname()
	if err != nil {
		hostname = err.Error()
		os.Exit(1)
	}
	log.Println(fmt.Sprintf("Host : %s, IP Server : %s, IP Request : %s", hostname, ipServer, IPAddress))

	return c.JSON(http.StatusOK, map[string]interface{}{
		"Data":      "Server is up and running",
		"IPRequest": IPAddress,
		"IPServer":  ipServer,
		"Hostname":  hostname,
	})
}

func externalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("are you connected to the network?")
}
