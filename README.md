#zipkin tracing & Wrapper


This is an example of how to integrate zipkin use  wrappers, a form of middleware,  
in go-micro

- main.go - the service with the handler wrappers
- cli/main.go - the client with a log client wrapper and zipkin tracing

#####be advise
微服务架构时一个业务流程往往需要调用多个微服务，这时也就需要整合链路追踪来追踪一次完成的跨服务  
调用，本文使用go-micro，用zipkin来实现tracing,cli之main.go模拟了三个微服务的调用，每个微服务的  
调用都生成了一个子childSpan。同时在服务提供端都会以Wrapper的形式来记录一次调用。