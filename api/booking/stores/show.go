package stores

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/imrenagi/concurrent-booking/api/booking"
)

func NewShow(db *gorm.DB) *Show {
	return &Show{
		db: db,
	}
}

type Show struct {
	db *gorm.DB
}

func (c Show) Booking(ctx context.Context, ID uuid.UUID) error {
	//
	// span := trace.SpanFromContext(ctx)
	// defer span.End()
	//
	// r := tc.NewRequest{
	// 	Requestid: "1",
	// 	TraceID:   span.SpanContext().TraceID().String(),
	// 	SpanID:    span.SpanContext().SpanID().String(),
	// }
	//
	// spanCtx, _ := tc.ConstructNewSpanContext(r)
	//
	// requestContext := context.Background()
	// requestContext = trace.ContextWithSpanContext(requestContext, spanCtx)
	//
	// // var requestInLoopSpan trace.Span
	// // childContext, requestInLoopSpan := otel.Tracer("inboundmessage").Start(requestContext, "requestInLoopSpan")
	//
	// fn := func(ctx context.Context) {
	//
	// 	ctx, span := otel.GetTracerProvider().Tracer("testxxxx").Start(ctx, "hello-span")
	// 	time.Sleep(1 * time.Second)
	// 	defer span.End()
	// }
	//
	// fn(requestContext)

	return c.db.Transaction(func(tx *gorm.DB) error {
		var show *booking.Show
		err := tx.WithContext(ctx).
			Clauses(clause.Locking{Strength: "NO KEY UPDATE"}).
			Where("id = ?", ID).
			First(&show).Error
		if err != nil {
			return err
		}

		if show.RemainingTickets-1 < 0 {
			return fmt.Errorf("ticket is not available")
		}
		show.RemainingTickets -= 1

		log.Debug().Int("remaining_tickets", show.RemainingTickets).Msg("remaining tickets have been decreased")

		return tx.WithContext(ctx).Save(&show).Error
	})
}

func (c Show) FindConcertByID(ctx context.Context, ID uuid.UUID) (*booking.Show, error) {
	var show *booking.Show
	err := c.db.WithContext(ctx).
		Where("id = ?", ID).
		First(&show).Error
	return show, err
}

func (c Show) Save(ctx context.Context, concert *booking.Show) error {
	return c.db.WithContext(ctx).Save(&concert).Error
}
