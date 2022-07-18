package telemetry

import (
	"fmt"

	"go.opentelemetry.io/otel/trace"
)

type NewRequest struct {
	Requestid string `json: "requestid"`
	TraceID   string
	SpanID    string
}

func ConstructNewSpanContext(request NewRequest) (spanContext trace.SpanContext, err error) {
	var traceID trace.TraceID
	traceID, err = trace.TraceIDFromHex(request.TraceID)
	if err != nil {
		fmt.Println("error: ", err)
		return spanContext, err
	}
	var spanID trace.SpanID
	spanID, err = trace.SpanIDFromHex(request.SpanID)
	if err != nil {
		fmt.Println("error: ", err)
		return spanContext, err
	}
	var spanContextConfig trace.SpanContextConfig
	spanContextConfig.TraceID = traceID
	spanContextConfig.SpanID = spanID
	spanContextConfig.TraceFlags = 01
	spanContextConfig.Remote = false
	spanContext = trace.NewSpanContext(spanContextConfig)
	return spanContext, nil
}
