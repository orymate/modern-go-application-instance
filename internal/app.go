package internal

import (
	"net/http"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/goph/emperror"
	"github.com/goph/logur"
	"github.com/gorilla/mux"
	"github.com/orymate/modern-go-application-instance/internal/greeting"
	"github.com/orymate/modern-go-application-instance/internal/greeting/greetingadapter"
	"github.com/orymate/modern-go-application-instance/internal/greeting/greetingdriver"
	"github.com/orymate/modern-go-application-instance/internal/greetingworker"
	"github.com/orymate/modern-go-application-instance/internal/greetingworker/greetingworkeradapter"
	"github.com/orymate/modern-go-application-instance/internal/greetingworker/greetingworkerdriver"
	"github.com/orymate/modern-go-application-instance/internal/httpbin"
)

// NewApp returns a new HTTP application.
func NewApp(logger logur.Logger, publisher message.Publisher, errorHandler emperror.Handler) http.Handler {
	helloWorld := greeting.NewHelloWorld(
		greetingadapter.NewHelloWorldEvents(publisher),
		greetingadapter.NewLogger(logger),
		errorHandler,
	)
	sayHello := greeting.NewSayHello(
		greetingadapter.NewSayHelloEvents(publisher),
		greetingadapter.NewLogger(logger),
		errorHandler,
	)
	helloWorldController := greetingdriver.NewGreetingController(helloWorld, sayHello, errorHandler)

	router := mux.NewRouter()

	router.Path("/").Methods("GET").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html")
		_, _ = w.Write([]byte(template))
	})

	router.Path("/hello").Methods("GET").HandlerFunc(helloWorldController.HelloWorld)
	router.Path("/hello").Methods("POST").HandlerFunc(helloWorldController.SayHello)

	router.PathPrefix("/httpbin").Handler(http.StripPrefix("/httpbin", httpbin.New()))

	return router
}

// RegisterEventHandlers registers event handlers in a message router.
func RegisterEventHandlers(router *message.Router, subscriber message.Subscriber, logger logur.Logger) error {
	helloWorldHandler := greetingworkerdriver.NewHelloWorldEventHandler(
		greetingworker.NewHelloWorldEventLogger(greetingworkeradapter.NewLogger(logger)),
	)

	err := router.AddNoPublisherHandler(
		"log_said_hello",
		"said_hello",
		subscriber,
		helloWorldHandler.SaidHello,
	)
	if err != nil {
		return err
	}

	sayHelloHandler := greetingworkerdriver.NewSayHelloEventHandler(
		greetingworker.NewSayHelloEventLogger(greetingworkeradapter.NewLogger(logger)),
	)

	err = router.AddNoPublisherHandler(
		"log_said_hello_to",
		"said_hello_to",
		subscriber,
		sayHelloHandler.SaidHelloTo,
	)
	if err != nil {
		return err
	}

	return nil
}
