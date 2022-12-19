package tap

import (
	"context"
	"io"

	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_v3_ext_proc_pb "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
	"github.com/safedep/gateway/services/pkg/common/logger"
	"github.com/safedep/gateway/services/pkg/common/messaging"
	"github.com/safedep/gateway/services/pkg/common/obs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	obsKeyTapReqType             = "tap_req_type"
	tapResponseTapSignatureKey   = "x-gateway-tap"
	tapResponseTapSignatureValue = "true"
)

type tapService struct {
	handlerChain     TapHandlerChain
	messagingService messaging.MessagingService
}

func NewTapService(msgService messaging.MessagingService,
	registrations []TapHandlerRegistration) (envoy_v3_ext_proc_pb.ExternalProcessorServer, error) {

	return &tapService{messagingService: msgService,
		handlerChain: TapHandlerChain{Handlers: registrations}}, nil
}

func (s *tapService) RegisterHandler(handler TapHandlerRegistration) {
	s.handlerChain.Handlers = append(s.handlerChain.Handlers, handler)
}

func (s *tapService) Process(srv envoy_v3_ext_proc_pb.ExternalProcessor_ProcessServer) error {
	logger.Debugf("Tap service: Handling stream")

	ctx := srv.Context()
	for {
		select {
		case <-ctx.Done():
			logger.Debugf("Context is finished: %v", ctx.Err())
			return ctx.Err()
		default:
		}

		req, err := srv.Recv()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			logger.Errorf("Received error from stream: %v", err)
			return status.Errorf(codes.Unknown, "Error receiving request: %v", err)
		}

		resp := &envoy_v3_ext_proc_pb.ProcessingResponse{}
		switch req.Request.(type) {
		case *envoy_v3_ext_proc_pb.ProcessingRequest_RequestHeaders:
			obs.SetAttributeInContext(ctx, obsKeyTapReqType,
				"ProcessingRequest_RequestHeaders")

			err = s.handleRequestHeaders(ctx,
				req.Request.(*envoy_v3_ext_proc_pb.ProcessingRequest_RequestHeaders))

			resp.Response = &envoy_v3_ext_proc_pb.ProcessingResponse_RequestHeaders{
				RequestHeaders: &envoy_v3_ext_proc_pb.HeadersResponse{
					Response: &envoy_v3_ext_proc_pb.CommonResponse{
						Status: envoy_v3_ext_proc_pb.CommonResponse_CONTINUE,
					},
				},
			}

			// TODO: Use handler chain for applying upstream auth
			err = s.applyUpstreamAuth(req.Request.(*envoy_v3_ext_proc_pb.ProcessingRequest_RequestHeaders),
				resp.Response.(*envoy_v3_ext_proc_pb.ProcessingResponse_RequestHeaders))
			break
		case *envoy_v3_ext_proc_pb.ProcessingRequest_ResponseHeaders:
			obs.SetAttributeInContext(ctx, obsKeyTapReqType,
				"ProcessingRequest_ResponseHeaders")

			err = s.handleResponseHeaders(ctx,
				req.Request.(*envoy_v3_ext_proc_pb.ProcessingRequest_ResponseHeaders))
			s.addTapSignature(resp)
			break
		default:
			logger.Warnf("Unknown request type: %v", req.Request)
		}

		// TODO: How should we handle this behavior?
		if err != nil {
			logger.Warnf("Error in handling processing req: %v", err)
		}

		if err = srv.Send(resp); err != nil {
			logger.Warnf("Failed to send stream response: %v", err)
		}
	}
}

func (s *tapService) handleRequestHeaders(ctx context.Context,
	req *envoy_v3_ext_proc_pb.ProcessingRequest_RequestHeaders) error {
	for _, registration := range s.handlerChain.Handlers {
		err := registration.Handler.HandleRequestHeaders(ctx, req)
		if !registration.ContinueOnError && err != nil {
			logger.Warnf("Unable to continue on tap handler error: %v", err)
			return err
		}
	}

	return nil
}

func (s *tapService) handleResponseHeaders(ctx context.Context,
	req *envoy_v3_ext_proc_pb.ProcessingRequest_ResponseHeaders) error {
	for _, registration := range s.handlerChain.Handlers {
		err := registration.Handler.HandleResponseHeaders(ctx, req)
		if !registration.ContinueOnError && err != nil {
			logger.Warnf("Unable to continue on tap handler error: %v", err)
			return err
		}
	}

	return nil
}

// Lets add a tap signature only if the response is not already used
func (s *tapService) addTapSignature(resp *envoy_v3_ext_proc_pb.ProcessingResponse) {
	if resp.Response != nil {
		return
	}

	logger.Debugf("Adding tap signature to response headers")
	resp.Response = &envoy_v3_ext_proc_pb.ProcessingResponse_ResponseHeaders{
		ResponseHeaders: &envoy_v3_ext_proc_pb.HeadersResponse{
			Response: &envoy_v3_ext_proc_pb.CommonResponse{
				HeaderMutation: &envoy_v3_ext_proc_pb.HeaderMutation{
					SetHeaders: []*envoy_config_core_v3.HeaderValueOption{
						{
							Header: &envoy_config_core_v3.HeaderValue{
								Key:   tapResponseTapSignatureKey,
								Value: tapResponseTapSignatureValue,
							},
						},
					},
				},
			},
		},
	}
}
