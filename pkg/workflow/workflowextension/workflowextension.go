package workflowextension

import (
	"reflect"
	"strings"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/event"

	"github.com/cloudevents/sdk-go/v2/types"
)

const (
	WorkflowIdCloudEventsExtension = "workflowid"

)

// WorkflowExtension represents the extension for extension context
type WorkflowExtension struct {
	WorkflowId string `json:"workflowid"`
}

// AddWorkflowAttributes adds the workflow attributes to the extension context
func (w WorkflowExtension) AddWorkflowAttributes(e event.EventWriter) {
	if w.WorkflowId != "" {
		value := reflect.ValueOf(w)
		typeOf := value.Type()

		for i := 0; i < value.NumField(); i++ {
			k := strings.ToLower(typeOf.Field(i).Name)
			v := value.Field(i).Interface()
			if k == WorkflowIdCloudEventsExtension && v == "" {
				continue
			}
			e.SetExtension(k, v)
		}
	}
}

func GetWorkflowExtension(event event.Event) (WorkflowExtension, bool) {
	if workflowExtension, ok := event.Extensions()[WorkflowIdCloudEventsExtension]; ok {
		if workflowExtensionString, err := types.ToString(workflowExtension); err == nil {

			return WorkflowExtension{WorkflowId: workflowExtensionString}, true
		}
	}
	return WorkflowExtension{}, false
}

func (w *WorkflowExtension) ReadTransformer() binding.TransformerFunc {
	return func(reader binding.MessageMetadataReader, writer binding.MessageMetadataWriter) error {
		workflowExtension := reader.GetExtension(WorkflowIdCloudEventsExtension)
		if workflowExtension != nil {
			tpFormatted, err := types.Format(workflowExtension)
			if err != nil {
				return err
			}
			w.WorkflowId = tpFormatted
		}
		return nil
	}
}

func (w *WorkflowExtension) WriteTransformer() binding.TransformerFunc {
	return func(reader binding.MessageMetadataReader, writer binding.MessageMetadataWriter) error {
		err := writer.SetExtension(WorkflowIdCloudEventsExtension, w.WorkflowId)
		if err != nil {
			return nil
		}
		return nil
	}
}