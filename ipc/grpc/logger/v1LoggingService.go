package logger

import (
	"context"
	"errors"
	"github.com/meschbach/go-junk-bucket/pkg/streams"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/types/known/timestamppb"
	"io"
	"sync"
)

type v1LoggingService struct {
	UnimplementedV1Server
	changes sync.Mutex
	facet   loggingFacet
	streams map[int64]*drain
	nextID  int64
}

func (v *v1LoggingService) RegisterDrain(ctx context.Context, request *RegisterDrainRequest) (*RegisterDrainReply, error) {
	outputBuffer := streams.NewBuffer[bufferedEntry](32)
	observatory := &observedSink[bufferedEntry]{
		target: outputBuffer,
	}
	success, err := v.facet.registerDrain(ctx, request.Name, observatory)
	if err != nil {
		return nil, err
	}
	if !success {
		return &RegisterDrainReply{DrainID: -1}, nil
	}

	//lock for next ID
	v.changes.Lock()
	defer v.changes.Unlock()
	id := v.nextID
	v.nextID++
	outputPort := &drain{
		id:          id,
		stream:      outputBuffer,
		observatory: observatory,
	}
	v.streams[id] = outputPort

	return &RegisterDrainReply{
		DrainID:       id,
		InitialOffset: 0,
	}, nil
}

func (v *v1LoggingService) ReadDrain(ctx context.Context, in *ReadDrainRequest) (out *ReadDrainReply, err error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.Int64("drain.id", in.DrainID))
	//todo: verify
	drain, has := v.streams[in.DrainID]
	span.SetAttributes(attribute.Bool("drain.has", has))
	if !has {
		return nil, nil
	}
	drain.changes.Lock()
	defer drain.changes.Unlock()

	output := make([]bufferedEntry, in.Count)
	count, err := drain.stream.ReadSlice(ctx, output)
	span.SetAttributes(attribute.Int("read.count", count))
	if err != nil {
		if errors.Is(err, streams.UnderRun) || errors.Is(err, streams.Done) {
			span.SetAttributes(attribute.Bool("eos", true))
			return &ReadDrainReply{
				Entries:         nil,
				BeginningOffset: drain.offset,
				NextOffset:      drain.offset,
			}, nil
		}
		span.SetStatus(codes.Error, "reading slice")
		span.RecordError(err)
		return nil, err
	}
	span.SetAttributes(attribute.Bool("eos", false))

	originalOffset := drain.offset
	nextOffset := originalOffset + int64(count)
	drain.offset = nextOffset

	out = &ReadDrainReply{
		Entries:         make([]*ReadDrainReply_LogEvent, count),
		BeginningOffset: originalOffset,
		NextOffset:      nextOffset,
	}
	for index, e := range output[0:count] {
		out.Entries[index] = &ReadDrainReply_LogEvent{
			Offset: originalOffset + int64(index),
			When:   timestamppb.New(e.when),
			Text:   e.text,
			Origin: &LogOrigin{
				Kind:    e.from.Type.Kind,
				Version: e.from.Type.Version,
				Name:    e.from.Name,
				Stream:  e.streamName,
			},
		}
	}
	return out, nil
}

func (v *v1LoggingService) WatchDrain(clientPipe V1_WatchDrainServer) error {
	span := trace.SpanFromContext(clientPipe.Context())
	req, err := clientPipe.Recv()
	if err != nil {
		span.SetStatus(codes.Error, "failed on initial read")
		span.RecordError(err)
		return err
	}
	if req.DrainID == nil || req.Close != nil {
		span.SetStatus(codes.Error, "protocol violation")
		return errors.New("on first message drain ID must not be nil and close must be nil")
	}

	v.changes.Lock()
	drain, has := v.streams[*req.DrainID]
	v.changes.Unlock()
	if !has {
		span.SetStatus(codes.Error, "no such drain")
		return errors.New("no such drain")
	}

	drain.changes.Lock()
	offset := drain.offset
	drain.changes.Unlock()

	sourceEvents := make(chan int64, 32)
	writeSub := drain.observatory.writtenEvent.On(func(ctx context.Context, event streams.Sink[bufferedEntry]) {
		offset++
		span.AddEvent("observatory.writtenEvent")
		sourceEvents <- offset
	})
	drain.observatory.SinkEvents().Finished.On(func(ctx context.Context, event streams.Sink[bufferedEntry]) {
		span.AddEvent("observatory.Finished")
		close(sourceEvents)
	})
	defer writeSub.Off()

	clientEvents := make(chan *WatchDrainRequest, 32)
	go func() {
		for {
			req, err := clientPipe.Recv()
			//todo: need to propagate and close
			if err != nil {
				//todo: what is the difference between the two branches?
				if errors.Is(err, io.EOF) || errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
					span.AddEvent("client.graceful")
					truePtr := true
					clientEvents <- &WatchDrainRequest{
						Close: &truePtr,
					}
					return
				} else {
					//todo: handle grpc errors: https://jbrandhorst.com/post/grpc-errors/
					span.AddEvent("client.error")
					span.SetStatus(codes.Error, "client receive error")
					span.RecordError(err)
					truePtr := true
					clientEvents <- &WatchDrainRequest{
						Close: &truePtr,
					}
					return
				}
			}
			//todo: really an error -- should probably forcibly close
			if req == nil {
				continue
			}
			if req.Close != nil && *req.Close {
				span.AddEvent("client.close-request")
				clientEvents <- req
				close(clientEvents)
				return
			}
		}
	}()

	if err := clientPipe.Send(&WatchDrainEvent{
		Offset: offset,
	}); err != nil {
		return err
	}
	for {
		select {
		case c, ok := <-clientEvents:
			span.AddEvent("client.close-closing")
			if !ok || (c.Close != nil && *c.Close) {
				//close(sourceEvents)
				return nil
			}
		case e := <-sourceEvents:
			if err := (func() error {
				_, span := tracing.Start(clientPipe.Context(), "V1LoggingService.SourceEvent")
				defer span.End()
				if err := clientPipe.Send(&WatchDrainEvent{
					Offset: e,
				}); err != nil {
					if err == context.Canceled {
						err = nil
					} else {
						span.SetStatus(codes.Error, "sending event")
						span.RecordError(err)
					}
				}
				return err
			})(); err != nil {
				return err
			}
		}
	}
}

func newV1LoggingService(facet loggingFacet) *v1LoggingService {
	return &v1LoggingService{
		changes: sync.Mutex{},
		facet:   facet,
		streams: make(map[int64]*drain),
		nextID:  1,
	}
}
