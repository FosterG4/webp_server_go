package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"webp_server_go/config"
	"webp_server_go/encoder"
	"webp_server_go/handler"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	log "github.com/sirupsen/logrus"
)

// https://docs.gofiber.io/api/fiber
var app = fiber.New(fiber.Config{
	ServerHeader:          "WebP Server Go",
	AppName:               "WebP Server Go",
	DisableStartupMessage: true,
	ProxyHeader:           "X-Real-IP",
	ReadBufferSize:        config.Config.ReadBufferSize,   // per-connection buffer size for requests' reading. This also limits the maximum header size. Increase this buffer if your clients send multi-KB RequestURIs and/or multi-KB headers (for example, BIG cookies).
	Concurrency:           config.Config.Concurrency,      // Maximum number of concurrent connections.
	DisableKeepalive:      config.Config.DisableKeepalive, // Disable keep-alive connections, the server will close incoming connections after sending the first response to the client
})

func setupLogger() {
	log.SetOutput(os.Stdout)
	log.SetReportCaller(true)
	formatter := &log.TextFormatter{
		EnvironmentOverrideColors: true,
		FullTimestamp:             true,
		TimestampFormat:           config.TimeDateFormat,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			return fmt.Sprintf("[%d:%s]", f.Line, f.Function), ""
		},
	}
	log.SetFormatter(formatter)
	log.SetLevel(log.InfoLevel)

	// fiber logger format
	app.Use(logger.New(logger.Config{
		Format:     config.FiberLogFormat,
		TimeFormat: config.TimeDateFormat,
	}))
	app.Use(recover.New(recover.Config{}))
	log.Infoln("WebP Server Go ready.")
}

func init() {
	// Our banner
	banner := fmt.Sprintf(`
		▌ ▌   ▌  ▛▀▖ ▞▀▖                ▞▀▖
		▌▖▌▞▀▖▛▀▖▙▄▘ ▚▄ ▞▀▖▙▀▖▌ ▌▞▀▖▙▀▖ ▌▄▖▞▀▖
		▙▚▌▛▀ ▌ ▌▌   ▖ ▌▛▀ ▌  ▐▐ ▛▀ ▌   ▌ ▌▌ ▌
		▘ ▘▝▀▘▀▀ ▘   ▝▀ ▝▀▘▘   ▘ ▝▀▘▘   ▝▀ ▝▀
		
		WebP Server Go - v%s
		Develop by WebP Server team. https://github.com/webp-sh`, config.Version)
	// main init is the last one to be called
	flag.Parse()
	// process cli params
	if config.DumpConfig {
		fmt.Println(config.SampleConfig)
		os.Exit(0)
	}
	if config.ShowVersion {
		fmt.Printf("\n %c[1;32m%s%c[0m\n\n", 0x1B, banner+"", 0x1B)
		os.Exit(0)
	}
	config.LoadConfig()
	fmt.Printf("\n %c[1;32m%s%c[0m\n\n", 0x1B, banner, 0x1B)
	setupLogger()
}

func main() {
	if config.Prefetch {
		go encoder.PrefetchImages()
	}
	app.Use(etag.New(etag.Config{
		Weak: true,
	}))

	listenAddress := config.Config.Host + ":" + config.Config.Port
	app.Get("/*", handler.Convert)

	fmt.Println("WebP Server Go is Running on http://" + listenAddress)

	_ = app.Listen(listenAddress)

}
