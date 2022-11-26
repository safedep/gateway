package models

import (
	"encoding/json"
	"time"

	"github.com/safedep/gateway/services/gen"
	event_api "github.com/safedep/gateway/services/gen"

	"github.com/safedep/gateway/services/pkg/common/utils"
)

func (m MetaEventWithAttributes) Serialize() ([]byte, error) {
	bytes, err := json.Marshal(m)
	if err != nil {
		return []byte{}, err
	} else {
		return bytes, nil
	}
}

func newMetaEventWithAttributes(t string) MetaEventWithAttributes {
	return MetaEventWithAttributes{
		MetaEvent: MetaEvent{
			Type:    t,
			Version: EventSchemaVersion,
		},
		MetaAttributes: MetaAttributes{},
	}
}

// Utils for new spec driven events
func eventUid() string {
	return utils.NewUniqueId()
}

func eventTimestamp(ts time.Time) *event_api.EventTimestamp {
	return &event_api.EventTimestamp{
		Seconds: ts.Unix(),
		Nanos:   int32(ts.Nanosecond()),
	}
}

func NewSpecEventHeader(tp event_api.EventType, source string) *event_api.EventHeader {
	return &event_api.EventHeader{
		Type:    tp,
		Source:  source,
		Id:      eventUid(),
		Context: &event_api.EventContext{},
	}
}

func NewSpecHeaderWithContext(tp event_api.EventType, source string, ctx *event_api.EventContext) *event_api.EventHeader {
	eh := NewSpecEventHeader(tp, source)
	eh.Context = ctx

	return eh
}

func NewArtefactRequestEvent(a Artefact, src string) *gen.TapArtefactRequestEvent {
	eh := NewSpecHeaderWithContext(event_api.EventType_TapArtefactReqEvent, src, &event_api.EventContext{})
	return &gen.TapArtefactRequestEvent{
		Header: eh,
		Data: &gen.TapArtefactRequestEvent_Data{
			Artefact: &gen.Artefact{
				Ecosystem: a.OpenSsfEcosystem(),
				Group:     a.Group,
				Name:      a.Name,
				Version:   a.Version,
			},
		},
		Timestamp: time.Now().UnixMilli(),
	}
}
