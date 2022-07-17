package middleware

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"
	"go.opentelemetry.io/otel/trace"

	"github.com/imrenagi/concurrent-booking/api/booking/services"
	"github.com/imrenagi/concurrent-booking/api/internal/telemetry"
)

func SpanPropagator(next asynq.Handler) asynq.Handler {
	return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
		var tp services.TaskPayload
		err := json.Unmarshal(t.Payload(), &tp)
		if err != nil {
			return err
		}
		spanCtx, _ := telemetry.ConstructNewSpanContext(telemetry.NewRequest{
			TraceID: tp.Context.TraceID,
			SpanID:  tp.Context.SpanID,
		})
		ctx = trace.ContextWithSpanContext(ctx, spanCtx)
		payload, err := json.Marshal(tp.Data)
		if err != nil {
			return err
		}
		return next.ProcessTask(ctx, asynq.NewTask(t.Type(), payload))
	})
}
