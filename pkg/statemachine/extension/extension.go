package extension

import (
	"reflect"
	"strings"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/event"

	"github.com/cloudevents/sdk-go/v2/types"
)

const (
	StateMachineIdCloudEventsExtension = "statemachineid"
	CorrelationKeyCloudEventsExtension = "correlationkey"
)

// StateMachineExtension represents the extension for extension context
type StateMachineExtension struct {
	StateMachineId string `json:"statemachineid"`
	CorrelationKey string `json:"correlationkey"`
}

// AddStateMachineAttributes adds the statemachine attributes to the extension context
func (sme StateMachineExtension) AddStateMachineAttributes(e event.EventWriter) {
	if sme.StateMachineId != "" {
		value := reflect.ValueOf(sme)
		typeOf := value.Type()

		for i := 0; i < value.NumField(); i++ {
			k := strings.ToLower(typeOf.Field(i).Name)
			v := value.Field(i).Interface()
			if k == StateMachineIdCloudEventsExtension && v == "" {
				continue
			}
			e.SetExtension(k, v)
		}
	}
}

func GetStateMachineExtension(event event.Event) (StateMachineExtension, bool) {
	if stateMachineExtension, ok := event.Extensions()[StateMachineIdCloudEventsExtension]; ok {
		if stateMachineExtensionString, err := types.ToString(stateMachineExtension); err == nil {

			return StateMachineExtension{StateMachineId: stateMachineExtensionString}, true
		}
	}
	return StateMachineExtension{}, false
}

func (sme *StateMachineExtension) ReadTransformer() binding.TransformerFunc {
	return func(reader binding.MessageMetadataReader, writer binding.MessageMetadataWriter) error {
		stateMachineIdExtension := reader.GetExtension(StateMachineIdCloudEventsExtension)
		if stateMachineIdExtension != nil {
			tpFormatted, err := types.Format(stateMachineIdExtension)
			if err != nil {
				return err
			}
			sme.StateMachineId = tpFormatted
		}

		stateMachineCorrelationKeyExtension := reader.GetExtension(CorrelationKeyCloudEventsExtension)
		if stateMachineCorrelationKeyExtension != nil {
			tpFormatted, err := types.Format(stateMachineCorrelationKeyExtension)
			if err != nil {
				return err
			}
			sme.CorrelationKey = tpFormatted
		}
		return nil
	}
}

func (sme *StateMachineExtension) WriteTransformer() binding.TransformerFunc {
	return func(reader binding.MessageMetadataReader, writer binding.MessageMetadataWriter) error {
		err := writer.SetExtension(StateMachineIdCloudEventsExtension, sme.StateMachineId)
		if err != nil {
			return nil
		}
		err2 := writer.SetExtension(CorrelationKeyCloudEventsExtension, sme.CorrelationKey)
		if err2 != nil {
			return nil
		}
		return nil
	}
}