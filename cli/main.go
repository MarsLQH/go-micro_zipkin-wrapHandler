package main

import (
	"fmt"
	helloProto "github.com/asim/go-micro/examples/v4/greeter/srv/proto/hello"
	zipkintrace "github.com/asim/go-micro/examples/v4/wrapper/trace"
	"github.com/opentracing/opentracing-go"
	opentracingLog "github.com/opentracing/opentracing-go/log"
	opentrcingZipkinImpl "github.com/openzipkin-contrib/zipkin-go-opentracing"
	"github.com/openzipkin/zipkin-go"
	httpReporter "github.com/openzipkin/zipkin-go/reporter/http"
	"go-micro.dev/v4/metadata"
	"io/ioutil"
	"log"
	"net/http"

	"context"
	proto "github.com/asim/go-micro/examples/v4/service/proto"
	"go-micro.dev/v4"
	"go-micro.dev/v4/client"
)

// log wrapper logs every time a request is made
type logWrapper struct {
	client.Client
}

func (l *logWrapper) Call(ctx context.Context, req client.Request, rsp interface{}, opts ...client.CallOption) error {
	fmt.Printf("[wrapper] client request service: %s method: %s\n", req.Service(), req.Endpoint())
	return l.Client.Call(ctx, req, rsp)
}

// Implements client.Wrapper as logWrapper
func logWrap(c client.Client) client.Client {
	return &logWrapper{c}
}

func main() {
	service := micro.NewService(
		micro.Name("greeter.client"),
		// wrap the client
		micro.WrapClient(logWrap),
	)

	service.Init()

	greeter := proto.NewGreeterService("go.micro.srv.wrapper", service.Client())

	rsp, err := greeter.Hello(context.TODO(), &proto.Request{Name: "John"})
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(rsp.Greeting)

	//=========

	reporter := httpReporter.NewReporter(zipkintrace.ZipkinHttpReportHost)
	defer reporter.Close()

	endpoint, err := zipkin.NewEndpoint("test-cli", "localhost:5999")
	if err != nil {
		log.Fatalln(err)
	}

	tracer, err := zipkin.NewTracer(
		reporter,
		zipkin.WithLocalEndpoint(endpoint),
		zipkin.WithSampler(zipkin.AlwaysSample),
	)

	if err != nil {
		log.Fatalln(err)
	}

	globalTracer := opentrcingZipkinImpl.Wrap(tracer)
	opentracing.SetGlobalTracer(globalTracer)

//==
	//zipkintrace.InitZipKinTrace("test-cli", "localhost:5999")
//	zipkintrace.InitTracing()
	//InitTracing()

	span := opentracing.StartSpan("cli-request")
	defer span.Finish()

	request, err := http.NewRequest(http.MethodGet, "http://localhost:6063/", nil)
	if err != nil {
		span.SetTag("error",  err.Error())
		span.LogFields(opentracingLog.Error(err))
	}

	err = opentracing.GlobalTracer().Inject(
		span.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(request.Header),
	)

	fmt.Println(request.Header)

	if err != nil {
		span.SetTag("error", err.Error())
		log.Fatalln(err)
	}

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		span.SetTag("error", err.Error())
		log.Fatalln(err)
	}

	childSpan := opentracing.StartSpan("request-finish", opentracing.ChildOf(span.Context()))
	defer childSpan.Finish()

	childSpan.LogFields(opentracingLog.Int("Status code", resp.StatusCode))
	content, err := ioutil.ReadAll(resp.Body)

	fmt.Println(string(content))

	//=======					==============
	cl := helloProto.NewSayService("go.micro.srv.metadata", service.Client())

	// Set arbitrary headers in context
	ctx := metadata.NewContext(context.Background(), map[string]string{
		"User": "MarsLuo",
		"ID":   "20160126-20191111-20211021",
	})

	childSpan2 := opentracing.StartSpan("metadata-cli-request", opentracing.ChildOf(span.Context()))
	defer childSpan2.Finish()

	rsp2, err2 := cl.Hello(ctx, &helloProto.Request{
		Name: "Jupiter",
	})
	if err2 != nil {
		fmt.Println(err)
		childSpan2.SetTag("error2",  err2.Error())
		childSpan2.LogFields(opentracingLog.Error(err2))
		return
	}
	childSpan2.SetTag("metadata rsp ok",  "ok")
	childSpan2.LogFields(opentracingLog.Object ("metadata rsp2 ", rsp2))
	fmt.Println(rsp2)
	//======================call another service
	childSpan3 := opentracing.StartSpan("helloworld-cli-request", opentracing.ChildOf(span.Context()))
	defer childSpan3.Finish()
	cl2:=	proto.NewGreeterService("helloworld",service.Client())
	rsp3, err2 :=cl2.Hello(ctx, &proto.Request{
		Name: "Saturn",
	} )
	if err2 != nil {
		fmt.Println(err2)
		childSpan3.SetTag("error3",  err2.Error())
		childSpan3.LogFields(opentracingLog.Error(err2))
		return
	}
	childSpan3.SetTag("helloworld rsp ok",  "ok")
	childSpan3.LogFields(opentracingLog.Object ("helloworld rsp3 ", rsp3))
	fmt.Println(rsp3)

}

func InitTracing()  {
	reporter := httpReporter.NewReporter( zipkintrace.ZipkinHttpReportHost)
	//defer reporter.Close()

	endpoint, err := zipkin.NewEndpoint("test-cli", "localhost:5999")
	if err != nil {
		log.Fatalln(err)
	}

	tracer, err := zipkin.NewTracer(
		reporter,
		zipkin.WithLocalEndpoint(endpoint),
		zipkin.WithSampler(zipkin.AlwaysSample),
	)

	if err != nil {
		log.Fatalln(err)
	}

	globalTracer := opentrcingZipkinImpl.Wrap(tracer)
	opentracing.SetGlobalTracer(globalTracer)
}