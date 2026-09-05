package trogonerror

import (
	"fmt"

	errpb "github.com/TrogonStack/trogonproto/gen/trogon/error/v1alpha1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type protoFieldSpec struct {
	key            string
	number         protoreflect.FieldNumber
	visibility     Visibility
	hasFixedValue  bool
	fixedValue     string
	hasDefault     bool
	defaultValue   string
}

// NewErrorTemplateFromProto builds an ErrorTemplate from a proto message type
// that carries the trogon.error.v1alpha1.Template message option.
//
// The descriptor is read once at template-construction time. Per-field
// FieldOptions are cached so subsequent FromProto calls do not re-walk the
// descriptor.
func NewErrorTemplateFromProto[T proto.Message](options ...TemplateOption) *ErrorTemplate {
	var zero T
	desc := zero.ProtoReflect().Descriptor()

	template := &ErrorTemplate{
		code:       CodeUnknown,
		visibility: VisibilityInternal,
	}

	if msgOpts, ok := proto.GetExtension(desc.Options(), errpb.E_Message).(*errpb.MessageOptions); ok && msgOpts != nil {
		applyProtoTemplate(template, msgOpts.GetTemplate())
	}

	template.fields = collectFieldSpecs(desc)
	for _, spec := range template.fields {
		if spec.hasFixedValue {
			if template.metadata == nil {
				template.metadata = Metadata{}
			}
			template.metadata[spec.key] = MetadataValue{
				value:      spec.fixedValue,
				visibility: spec.visibility,
			}
		}
	}

	for _, option := range options {
		option(template)
	}

	return template
}

// FromProto creates a new error instance, deriving metadata from the proto
// message's populated fields according to their FieldOptions annotations.
//
// Fields with value_policy = value contribute their fixed literal (already
// baked into the template). Fields with value_policy = default_value use the
// runtime field value when set, falling back to default_value otherwise.
// Fields without a value policy use the runtime field value as-is.
//
// Caller-supplied options apply last and override anything derived from the
// proto instance.
func (et *ErrorTemplate) FromProto(m proto.Message, options ...ErrorOption) *TrogonError {
	derived := make([]ErrorOption, 0, len(et.fields))
	reflected := m.ProtoReflect()

	for _, spec := range et.fields {
		if spec.hasFixedValue {
			continue
		}

		field := reflected.Descriptor().Fields().ByNumber(spec.number)
		if field == nil {
			continue
		}

		value := protoFieldString(reflected, field)
		if value == "" && spec.hasDefault {
			value = spec.defaultValue
		}
		if value == "" {
			continue
		}

		derived = append(derived, WithMetadataValue(spec.visibility, spec.key, value))
	}

	return et.NewError(append(derived, options...)...)
}

func applyProtoTemplate(template *ErrorTemplate, t *errpb.MessageOptions_Template) {
	if t == nil {
		return
	}
	if d := t.GetDomain(); d != "" {
		template.domain = d
	}
	if r := t.GetReason(); r != "" {
		template.reason = r
	}
	if m := t.GetMessage(); m != "" {
		template.message = m
	}
	if c := t.GetCode(); c != errpb.Code_UNSPECIFIED {
		template.code = mapProtoCode(c)
	}
	if v := t.GetVisibility(); v != errpb.Visibility_VISIBILITY_UNSPECIFIED {
		template.visibility = mapProtoVisibility(v)
	}
	for _, link := range t.GetHelpLinks() {
		if template.help == nil {
			template.help = &Help{}
		}
		template.help.links = append(template.help.links, HelpLink{
			description: link.GetDescription(),
			url:         link.GetUrl(),
		})
	}
	for _, entry := range t.GetMetadata() {
		if entry.GetKey() == "" {
			continue
		}
		if template.metadata == nil {
			template.metadata = Metadata{}
		}
		template.metadata[entry.GetKey()] = MetadataValue{
			value:      entry.GetValue(),
			visibility: mapProtoVisibility(entry.GetVisibility()),
		}
	}
}

func collectFieldSpecs(desc protoreflect.MessageDescriptor) []protoFieldSpec {
	fields := desc.Fields()
	specs := make([]protoFieldSpec, 0, fields.Len())
	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		fopts, ok := proto.GetExtension(field.Options(), errpb.E_Field).(*errpb.FieldOptions)
		if !ok || fopts == nil {
			continue
		}
		spec := protoFieldSpec{
			key:        field.JSONName(),
			number:     field.Number(),
			visibility: mapProtoVisibility(fopts.GetVisibility()),
		}
		switch policy := fopts.GetValuePolicy().(type) {
		case *errpb.FieldOptions_Value:
			spec.hasFixedValue = true
			spec.fixedValue = policy.Value
		case *errpb.FieldOptions_DefaultValue:
			spec.hasDefault = true
			spec.defaultValue = policy.DefaultValue
		}
		specs = append(specs, spec)
	}
	return specs
}

func protoFieldString(m protoreflect.Message, field protoreflect.FieldDescriptor) string {
	if !m.Has(field) {
		return ""
	}
	v := m.Get(field)
	switch field.Kind() {
	case protoreflect.StringKind:
		return v.String()
	case protoreflect.EnumKind:
		if ev := field.Enum().Values().ByNumber(v.Enum()); ev != nil {
			return string(ev.Name())
		}
		return fmt.Sprint(int32(v.Enum()))
	default:
		return fmt.Sprint(v.Interface())
	}
}

func mapProtoCode(c errpb.Code) Code {
	switch c {
	case errpb.Code_CANCELLED:
		return CodeCancelled
	case errpb.Code_UNKNOWN:
		return CodeUnknown
	case errpb.Code_INVALID_ARGUMENT:
		return CodeInvalidArgument
	case errpb.Code_DEADLINE_EXCEEDED:
		return CodeDeadlineExceeded
	case errpb.Code_NOT_FOUND:
		return CodeNotFound
	case errpb.Code_ALREADY_EXISTS:
		return CodeAlreadyExists
	case errpb.Code_PERMISSION_DENIED:
		return CodePermissionDenied
	case errpb.Code_RESOURCE_EXHAUSTED:
		return CodeResourceExhausted
	case errpb.Code_FAILED_PRECONDITION:
		return CodeFailedPrecondition
	case errpb.Code_ABORTED:
		return CodeAborted
	case errpb.Code_OUT_OF_RANGE:
		return CodeOutOfRange
	case errpb.Code_UNIMPLEMENTED:
		return CodeUnimplemented
	case errpb.Code_INTERNAL:
		return CodeInternal
	case errpb.Code_UNAVAILABLE:
		return CodeUnavailable
	case errpb.Code_DATA_LOSS:
		return CodeDataLoss
	case errpb.Code_UNAUTHENTICATED:
		return CodeUnauthenticated
	default:
		return CodeUnknown
	}
}

func mapProtoVisibility(v errpb.Visibility) Visibility {
	switch v {
	case errpb.Visibility_VISIBILITY_PUBLIC:
		return VisibilityPublic
	case errpb.Visibility_VISIBILITY_PRIVATE:
		return VisibilityPrivate
	default:
		return VisibilityInternal
	}
}
