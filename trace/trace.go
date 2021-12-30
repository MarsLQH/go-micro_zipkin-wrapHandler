package zipkintrace

import (
	"context"
	"errors"
	"fmt"
	"github.com/opentracing/opentracing-go"
	opentracingLog "github.com/opentracing/opentracing-go/log"
	opentrcingZipkinImpl "github.com/openzipkin-contrib/zipkin-go-opentracing"
	"github.com/openzipkin/zipkin-go"
	httpReporter "github.com/openzipkin/zipkin-go/reporter/http"
	"go-micro.dev/v4/server"
	"log"
	"math/rand"
	"net/http"

	"time"
)

const ZipkinHttpReportHost = "http://127.0.0.0:9411/api/v2/spans"

func InitZipKinTrace(serverName string, hostPort string) {
	reporter := httpReporter.NewReporter(ZipkinHttpReportHost)
	fmt.Println( ZipkinHttpReportHost)
	defer reporter.Close()
	localEndpoint, err := zipkin.NewEndpoint(serverName, hostPort)
	if err != nil {
		log.Fatalln(err)
	}

	tracer, err := zipkin.NewTracer(
		reporter,
		zipkin.WithSampler(zipkin.AlwaysSample),
		zipkin.WithLocalEndpoint(localEndpoint),
	)
	if err != nil {
		log.Fatalln(err)
	}

	globalTracer := opentrcingZipkinImpl.Wrap(tracer)
	opentracing.SetGlobalTracer(globalTracer)

}

func TraceWrappers(fn server.HandlerFunc) server.HandlerFunc {

	return func(ctx context.Context, req server.Request, rsp interface{}) error {
		log.Printf("[====	It is from  Trace	==== ] server request: %v", req.Endpoint())
		var httpReq http.Request
		parentContext, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(httpReq.Header))
		var span opentracing.Span
		if err == nil {
			span = opentracing.StartSpan(req.Endpoint(), opentracing.ChildOf(parentContext))
		} else {
			span = opentracing.StartSpan(req.Endpoint())
		}
		defer span.Finish()

		span.SetTag("db-mysql", "localhost:3306")
		time.Sleep(time.Millisecond * 5)
		span.LogFields(opentracingLog.Int64("query-start", time.Now().Unix()))
		time.Sleep(time.Duration(rand.Intn(3)))
		span.LogFields(opentracingLog.Int64("query-end", time.Now().Unix()))

		span.LogFields(opentracingLog.Error(errors.New("Msql query failed 2")))
		span.SetTag("error", "Msql query failed tag2.")
		log.Printf("[Trace Wrapper 2] Before serving request method: %v\n", req.Method())
		err = fn(ctx, req, rsp)
		log.Printf("[Trace Wrapper 2] After serving request. TraceId: %v\n", opentracing.GlobalTracer())
		return err
	}
}
