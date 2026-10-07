// Package errproto derives trogonerror.ErrorTemplate values from proto
// messages that carry the trogon.error.v1alpha1 Template and FieldOptions
// message/field options, so services can declare their error contract once
// in proto and pick it up in Go without hand-rewriting the same
// domain/reason/code/visibility/help/metadata in two places.
package errproto

import (
	"fmt"

	"github.com/TrogonStack/trogonerror"
	errpb "github.com/TrogonStack/trogonproto/gen/trogon/error/v1alpha1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type protoFieldSpec struct {
	key           string
	number        protoreflect.FieldNumber
	visibility    trogonerror.Visibility
	hasFixedValue bool
	fixedValue    string
	hasDefault    bool
	defaultValue  string
}

// Template wraps a trogonerror.ErrorTemplate with the proto field
// specifications needed to derive per-instance metadata from a populated
// proto message.
type Template struct {
	*trogonerror.ErrorTemplate
	fields []protoFieldSpec
}

// NewErrorTemplateFromProto builds a Template from a proto message type
// that carries the trogon.error.v1alpha1.Template message option.
//
// The descriptor is read once at template-construction time. Per-field
// FieldOptions are cached so subsequent FromProto calls do not re-walk the
// descriptor.
func NewErrorTemplateFromProto[T proto.Message](options ...trogonerror.TemplateOption) *Template {
	var zero T
	desc := zero.ProtoReflect().Descriptor()

	var domain, reason string
	var derived []trogonerror.TemplateOption

	if msgOpts, ok := proto.GetExtension(desc.Options(), errpb.E_Message).(*errpb.MessageOptions); ok && msgOpts != nil {
		domain, reason, derived = templateOptionsFromProto(msgOpts.GetTemplate())
	}

	fields := collectFieldSpecs(desc)
	for _, spec := range fields {
		if spec.hasFixedValue {
			derived = append(derived, trogonerror.TemplateWithMetadataValue(spec.visibility, spec.key, spec.fixedValue))
		}
	}

	et := trogonerror.NewErrorTemplate(domain, reason, append(derived, options...)...)

	return &Template{ErrorTemplate: et, fields: fields}
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
func (t *Template) FromProto(m proto.Message, options ...trogonerror.ErrorOption) *trogonerror.TrogonError {
	derived := make([]trogonerror.ErrorOption, 0, len(t.fields))
	reflected := m.ProtoReflect()

	for _, spec := range t.fields {
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

		derived = append(derived, trogonerror.WithMetadataValue(spec.visibility, spec.key, value))
	}

	return t.NewError(append(derived, options...)...)
}

func templateOptionsFromProto(t *errpb.MessageOptions_Template) (domain, reason string, opts []trogonerror.TemplateOption) {
	if t == nil {
		return "", "", nil
	}

	domain = t.GetDomain()
	reason = t.GetReason()

	if m := t.GetMessage(); m != "" {
		opts = append(opts, trogonerror.TemplateWithMessage(m))
	}
	if c := t.GetCode(); c != errpb.Code_UNSPECIFIED {
		opts = append(opts, trogonerror.TemplateWithCode(mapProtoCode(c)))
	}
	if v := t.GetVisibility(); v != errpb.Visibility_VISIBILITY_UNSPECIFIED {
		opts = append(opts, trogonerror.TemplateWithVisibility(mapProtoVisibility(v)))
	}
	for _, link := range t.GetHelpLinks() {
		opts = append(opts, trogonerror.TemplateWithHelpLink(link.GetDescription(), link.GetUrl()))
	}
	for _, entry := range t.GetMetadata() {
		if entry.GetKey() == "" {
			continue
		}
		opts = append(opts, trogonerror.TemplateWithMetadataValue(mapProtoVisibility(entry.GetVisibility()), entry.GetKey(), entry.GetValue()))
	}

	return domain, reason, opts
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

func mapProtoCode(c errpb.Code) trogonerror.Code {
	switch c {
	case errpb.Code_CANCELLED:
		return trogonerror.CodeCancelled
	case errpb.Code_UNKNOWN:
		return trogonerror.CodeUnknown
	case errpb.Code_INVALID_ARGUMENT:
		return trogonerror.CodeInvalidArgument
	case errpb.Code_DEADLINE_EXCEEDED:
		return trogonerror.CodeDeadlineExceeded
	case errpb.Code_NOT_FOUND:
		return trogonerror.CodeNotFound
	case errpb.Code_ALREADY_EXISTS:
		return trogonerror.CodeAlreadyExists
	case errpb.Code_PERMISSION_DENIED:
		return trogonerror.CodePermissionDenied
	case errpb.Code_RESOURCE_EXHAUSTED:
		return trogonerror.CodeResourceExhausted
	case errpb.Code_FAILED_PRECONDITION:
		return trogonerror.CodeFailedPrecondition
	case errpb.Code_ABORTED:
		return trogonerror.CodeAborted
	case errpb.Code_OUT_OF_RANGE:
		return trogonerror.CodeOutOfRange
	case errpb.Code_UNIMPLEMENTED:
		return trogonerror.CodeUnimplemented
	case errpb.Code_INTERNAL:
		return trogonerror.CodeInternal
	case errpb.Code_UNAVAILABLE:
		return trogonerror.CodeUnavailable
	case errpb.Code_DATA_LOSS:
		return trogonerror.CodeDataLoss
	case errpb.Code_UNAUTHENTICATED:
		return trogonerror.CodeUnauthenticated
	default:
		return trogonerror.CodeUnknown
	}
}

func mapProtoVisibility(v errpb.Visibility) trogonerror.Visibility {
	switch v {
	case errpb.Visibility_VISIBILITY_PUBLIC:
		return trogonerror.VisibilityPublic
	case errpb.Visibility_VISIBILITY_PRIVATE:
		return trogonerror.VisibilityPrivate
	default:
		return trogonerror.VisibilityInternal
	}
}
