package binder

import (
	"bytes"
	"context"
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aatuh/pureapi-core/server"
	frameworkcontext "github.com/aatuh/pureapi-framework/context"
)

// Binder converts HTTP requests into typed inputs.
type Binder interface {
	Bind(ctx context.Context, r *http.Request, dest any) error
}

// BodyDecoder decodes a raw body into dest.
type BodyDecoder interface {
	Decode(data []byte, dest any) error
}

// JSONBodyDecoder implements BodyDecoder using encoding/json.
type JSONBodyDecoder struct {
	DisallowUnknown bool
}

// Decode satisfies BodyDecoder.
func (d JSONBodyDecoder) Decode(data []byte, dest any) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	if d.DisallowUnknown {
		dec.DisallowUnknownFields()
	}
	if err := dec.Decode(dest); err != nil {
		return fmt.Errorf("json decode: %w", err)
	}
	// Ensure no trailing tokens.
	if dec.More() {
		return fmt.Errorf("json decode: trailing data")
	}
	return nil
}

const defaultMaxBodyBytes int64 = 1 << 20 // 1MB

// DefaultBinder is the framework binder implementation.
type DefaultBinder struct {
	MaxBodyBytes     int64
	BodyDecoder      BodyDecoder
	ReadTimeout      time.Duration
	StrictJSONBodies bool
}

func NewDefaultBinder() *DefaultBinder {
	return &DefaultBinder{
		MaxBodyBytes:     defaultMaxBodyBytes,
		BodyDecoder:      JSONBodyDecoder{DisallowUnknown: false},
		ReadTimeout:      5 * time.Second,
		StrictJSONBodies: false,
	}
}

// WithStrictJSONBodies returns a copy with strict JSON body decoding toggled.
func (b *DefaultBinder) WithStrictJSONBodies(strict bool) *DefaultBinder {
	copy := *b
	copy.StrictJSONBodies = strict
	return &copy
}

// SetStrictJSONBodies toggles strict JSON decoding in place.
func (b *DefaultBinder) SetStrictJSONBodies(strict bool) {
	b.StrictJSONBodies = strict
}

func (b *DefaultBinder) bodyDecoder() BodyDecoder {
	if b.BodyDecoder == nil {
		return JSONBodyDecoder{DisallowUnknown: b.StrictJSONBodies}
	}
	switch dec := b.BodyDecoder.(type) {
	case JSONBodyDecoder:
		if b.StrictJSONBodies {
			dec.DisallowUnknown = true
		}
		return dec
	case *JSONBodyDecoder:
		if dec == nil {
			return JSONBodyDecoder{DisallowUnknown: b.StrictJSONBodies}
		}
		if b.StrictJSONBodies {
			dec.DisallowUnknown = true
		}
		return dec
	default:
		return b.BodyDecoder
	}
}

// FieldSource identifies where a value originated.
type FieldSource string

const (
	SourcePath   FieldSource = "path"
	SourceQuery  FieldSource = "query"
	SourceHeader FieldSource = "header"
	SourceCookie FieldSource = "cookie"
	SourceBody   FieldSource = "body"
)

// FieldError describes a single binding failure.
type FieldError struct {
	Field   string      `json:"field"`
	Source  FieldSource `json:"source"`
	Message string      `json:"message"`
}

// NewFieldError is a helper for constructing field-level errors.
func NewFieldError(field string, source FieldSource, message string) FieldError {
	return FieldError{
		Field:   field,
		Source:  source,
		Message: message,
	}
}

// BindError aggregates binding failures.
type BindError struct {
	message string
	fields  []FieldError
	cause   error
}

// NewBindError constructs a BindError with the provided message and fields.
func NewBindError(message string, fields []FieldError) *BindError {
	copyFields := make([]FieldError, len(fields))
	copy(copyFields, fields)
	return &BindError{
		message: message,
		fields:  copyFields,
	}
}

// WithCause returns a copy of the bind error with the cause attached.
func (e *BindError) WithCause(cause error) *BindError {
	if e == nil {
		return nil
	}
	clone := *e
	clone.cause = cause
	return &clone
}

// WithFields appends additional field errors and returns a new BindError.
func (e *BindError) WithFields(fields ...FieldError) *BindError {
	if e == nil {
		return nil
	}
	clone := *e
	clone.fields = append(append([]FieldError{}, clone.fields...), fields...)
	return &clone
}

// Message returns the human-readable message, if any.
func (e *BindError) Message() string {
	if e == nil {
		return ""
	}
	return e.message
}

// Error implements error.
func (e *BindError) Error() string {
	if e == nil {
		return ""
	}
	if e.message != "" {
		return e.message
	}
	if len(e.fields) == 0 {
		return "binding failed"
	}
	return fmt.Sprintf("binding failed: %s", e.fields[0].Message)
}

// Unwrap exposes the underlying cause, if any.
func (e *BindError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.cause
}

// Fields returns a copy of the field errors.
func (e *BindError) Fields() []FieldError {
	if e == nil {
		return nil
	}
	out := make([]FieldError, len(e.fields))
	copy(out, e.fields)
	return out
}

// WireMessage customizes the wire error message.
func (e *BindError) WireMessage() string {
	if e == nil {
		return ""
	}
	if e.message != "" {
		return e.message
	}
	return "Request could not be processed"
}

// WireData exposes detailed binding failures.
func (e *BindError) WireData() any {
	if e == nil {
		return nil
	}
	if len(e.fields) == 0 {
		return nil
	}
	return map[string]any{"fields": e.Fields()}
}

// ErrBodyTooLarge is returned when the request body exceeds MaxBodyBytes.
var ErrBodyTooLarge = errors.New("request body exceeds binder limit")

var textUnmarshalerType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()

// Bind populates dest based on struct tags.
func (b *DefaultBinder) Bind(ctx context.Context, r *http.Request, dest any) error {
	if ctx == nil {
		ctx = r.Context()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if dest == nil {
		return fmt.Errorf("binder destination must not be nil")
	}
	rv := reflect.ValueOf(dest)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return fmt.Errorf("binder destination must be a non-nil pointer")
	}
	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return fmt.Errorf("binder destination must point to a struct")
	}

	info := requestInfo{
		request:    r,
		pathParams: collectPathParams(r),
		query:      collectQueryValues(r),
		cookies:    collectCookies(r),
	}

	var bodyOnce sync.Once
	var bodyData []byte
	var bodyErr error
	getBody := func() ([]byte, error) {
		bodyOnce.Do(func() {
			bodyData, bodyErr = b.readBody(ctx, r)
		})
		return bodyData, bodyErr
	}

	var fieldErrors []FieldError
	if err := b.bindStruct(ctx, rv, "", info, &fieldErrors, getBody); err != nil {
		return err
	}
	if len(fieldErrors) > 0 {
		return &BindError{
			message: "Request failed validation",
			fields:  fieldErrors,
		}
	}
	return nil
}

type requestInfo struct {
	request    *http.Request
	pathParams map[string]string
	query      map[string][]string
	cookies    map[string]string
}

type bodyLoader func() ([]byte, error)

func (b *DefaultBinder) bindStruct(
	ctx context.Context,
	rv reflect.Value,
	parent string,
	info requestInfo,
	fieldErrors *[]FieldError,
	getBody bodyLoader,
) error {
	rt := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		fieldType := rt.Field(i)
		if !field.CanSet() {
			continue
		}
		name := fieldType.Name
		if parent != "" {
			name = parent + "." + name
		}

		// Anonymous struct fields cascade.
		if fieldType.Anonymous && field.Kind() == reflect.Struct && !hasBindingTag(fieldType) {
			if err := b.bindStruct(ctx, field, parent, info, fieldErrors, getBody); err != nil {
				return err
			}
			continue
		}

		if source, ok := fieldType.Tag.Lookup("path"); ok {
			key := firstNonEmpty(source, fieldType.Name)
			val, found := info.pathParams[key]
			if !found {
				if required(fieldType) {
					appendFieldError(fieldErrors, FieldError{Field: name, Source: SourcePath, Message: "missing required value"})
				}
				continue
			}
			if err := assignFromStrings(field, []string{val}); err != nil {
				appendFieldError(fieldErrors, FieldError{Field: name, Source: SourcePath, Message: err.Error()})
			}
			continue
		}

		if source, ok := fieldType.Tag.Lookup("query"); ok {
			key := firstNonEmpty(source, fieldType.Name)
			values := info.query[key]
			if len(values) == 0 {
				if required(fieldType) {
					appendFieldError(fieldErrors, FieldError{Field: name, Source: SourceQuery, Message: "missing required value"})
				}
				continue
			}
			if err := assignFromStrings(field, values); err != nil {
				appendFieldError(fieldErrors, FieldError{Field: name, Source: SourceQuery, Message: err.Error()})
			}
			continue
		}

		if source, ok := fieldType.Tag.Lookup("header"); ok {
			key := http.CanonicalHeaderKey(firstNonEmpty(source, fieldType.Name))
			values := info.request.Header.Values(key)
			if len(values) == 0 {
				if required(fieldType) {
					appendFieldError(fieldErrors, FieldError{Field: name, Source: SourceHeader, Message: "missing required value"})
				}
				continue
			}
			if err := assignFromStrings(field, values); err != nil {
				appendFieldError(fieldErrors, FieldError{Field: name, Source: SourceHeader, Message: err.Error()})
			}
			continue
		}

		if source, ok := fieldType.Tag.Lookup("cookie"); ok {
			key := firstNonEmpty(source, fieldType.Name)
			if val, ok := info.cookies[key]; ok {
				if err := assignFromStrings(field, []string{val}); err != nil {
					appendFieldError(fieldErrors, FieldError{Field: name, Source: SourceCookie, Message: err.Error()})
				}
			} else if required(fieldType) {
				appendFieldError(fieldErrors, FieldError{Field: name, Source: SourceCookie, Message: "missing required value"})
			}
			continue
		}

		if _, ok := fieldType.Tag.Lookup("body"); ok {
			data, err := getBody()
			if err != nil {
				if errors.Is(err, ErrBodyTooLarge) {
					return &BindError{
						message: "Request body too large",
						fields: []FieldError{
							{Field: name, Source: SourceBody, Message: "body size exceeds limit"},
						},
						cause: err,
					}
				}
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					return err
				}
				return &BindError{
					message: "Failed to read request body",
					fields:  []FieldError{NewFieldError(name, SourceBody, err.Error())},
					cause:   err,
				}
			}
			if len(data) == 0 {
				if required(fieldType) {
					appendFieldError(fieldErrors, FieldError{Field: name, Source: SourceBody, Message: "missing required value"})
				}
				continue
			}
			target := field
			if target.Kind() != reflect.Pointer {
				target = target.Addr()
			} else if target.IsNil() {
				target.Set(reflect.New(target.Type().Elem()))
			}
			if err := b.bodyDecoder().Decode(data, target.Interface()); err != nil {
				if msg := unknownJSONFieldMessage(err); msg != "" {
					return &BindError{
						message: "Unknown field in request body",
						fields:  []FieldError{NewFieldError(name, SourceBody, msg)},
						cause:   err,
					}
				}
				return &BindError{
					message: "Failed to decode request body",
					fields:  []FieldError{NewFieldError(name, SourceBody, err.Error())},
					cause:   err,
				}
			}
			if field.Kind() == reflect.Pointer {
				field.Set(target)
			}
			continue
		}

		if field.Kind() == reflect.Struct && fieldType.Tag == "" {
			if err := b.bindStruct(ctx, field, name, info, fieldErrors, getBody); err != nil {
				return err
			}
			continue
		}
	}
	return nil
}

func collectPathParams(r *http.Request) map[string]string {
	if params := server.RouteParams(r); params != nil {
		return params
	}
	if params := frameworkcontext.PathParamsFromContext(r.Context()); params != nil {
		return params
	}
	return map[string]string{}
}

func collectQueryValues(r *http.Request) map[string][]string {
	if qm := server.QueryMap(r); qm != nil {
		out := make(map[string][]string, len(qm))
		for k, v := range qm {
			switch vv := v.(type) {
			case string:
				out[k] = []string{vv}
			case []string:
				out[k] = append([]string{}, vv...)
			default:
				out[k] = []string{fmt.Sprint(v)}
			}
		}
		return out
	}
	values := r.URL.Query()
	out := make(map[string][]string, len(values))
	for k, v := range values {
		out[k] = append([]string{}, v...)
	}
	return out
}

func collectCookies(r *http.Request) map[string]string {
	out := make(map[string]string)
	for _, c := range r.Cookies() {
		out[c.Name] = c.Value
	}
	return out
}

func appendFieldError(dst *[]FieldError, fe FieldError) {
	*dst = append(*dst, fe)
}

func required(field reflect.StructField) bool {
	req, ok := field.Tag.Lookup("required")
	if !ok {
		return false
	}
	req = strings.TrimSpace(strings.ToLower(req))
	return req == "true" || req == "1" || req == "yes"
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func hasBindingTag(field reflect.StructField) bool {
	tags := []string{"path", "query", "header", "cookie", "body"}
	for _, tag := range tags {
		if _, ok := field.Tag.Lookup(tag); ok {
			return true
		}
	}
	return false
}

func assignFromStrings(field reflect.Value, values []string) error {
	if field.Kind() == reflect.Pointer {
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		return assignFromStrings(field.Elem(), values)
	}

	if field.Kind() == reflect.Slice {
		elemType := field.Type().Elem()
		slice := reflect.MakeSlice(field.Type(), 0, len(values))
		for _, v := range values {
			val, err := convertString(v, elemType)
			if err != nil {
				return err
			}
			slice = reflect.Append(slice, val)
		}
		field.Set(slice)
		return nil
	}

	if len(values) == 0 {
		return fmt.Errorf("no value provided")
	}

	val, err := convertString(values[0], field.Type())
	if err != nil {
		return err
	}
	field.Set(val)
	return nil
}

func convertString(input string, typ reflect.Type) (reflect.Value, error) {
	switch typ.Kind() {
	case reflect.String:
		return reflect.ValueOf(input).Convert(typ), nil
	case reflect.Bool:
		b, err := strconv.ParseBool(input)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("expected boolean")
		}
		return reflect.ValueOf(b).Convert(typ), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		bitSize := typ.Bits()
		i, err := strconv.ParseInt(input, 10, bitSize)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("expected integer")
		}
		v := reflect.New(typ).Elem()
		v.SetInt(i)
		return v, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		bitSize := typ.Bits()
		u, err := strconv.ParseUint(input, 10, bitSize)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("expected unsigned integer")
		}
		v := reflect.New(typ).Elem()
		v.SetUint(u)
		return v, nil
	case reflect.Float32, reflect.Float64:
		bitSize := typ.Bits()
		f, err := strconv.ParseFloat(input, bitSize)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("expected float")
		}
		v := reflect.New(typ).Elem()
		v.SetFloat(f)
		return v, nil
	case reflect.Struct:
		if typ == reflect.TypeOf(time.Time{}) {
			t, err := time.Parse(time.RFC3339, input)
			if err != nil {
				return reflect.Value{}, fmt.Errorf("expected RFC3339 timestamp")
			}
			return reflect.ValueOf(t), nil
		}
		if typ.Implements(textUnmarshalerType) {
			target := reflect.New(typ).Interface().(encoding.TextUnmarshaler)
			if err := target.UnmarshalText([]byte(input)); err != nil {
				return reflect.Value{}, err
			}
			return reflect.ValueOf(target).Elem(), nil
		}
	}

	if reflect.PointerTo(typ).Implements(textUnmarshalerType) {
		target := reflect.New(typ).Interface().(encoding.TextUnmarshaler)
		if err := target.UnmarshalText([]byte(input)); err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(target).Elem(), nil
	}

	return reflect.Value{}, fmt.Errorf("unsupported type %s", typ.String())
}

func (b *DefaultBinder) readBody(ctx context.Context, r *http.Request) ([]byte, error) {
	if r.Body == nil {
		return nil, nil
	}
	limit := b.MaxBodyBytes
	if limit <= 0 {
		limit = defaultMaxBodyBytes
	}
	if limit >= math.MaxInt64-1 {
		limit = math.MaxInt64 - 1
	}
	reader := io.LimitReader(r.Body, limit+1)
	data, err := readAllWithContext(ctx, reader, b.ReadTimeout)
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, ErrBodyTooLarge
	}
	// draining ensures body is fully read
	io.Copy(io.Discard, r.Body)
	r.Body.Close()
	return data, nil
}

type readResult struct {
	data []byte
	err  error
}

func readAllWithContext(ctx context.Context, reader io.Reader, timeout time.Duration) ([]byte, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	readCtx := ctx
	var cancel context.CancelFunc
	if timeout > 0 {
		readCtx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}
	ch := make(chan readResult, 1)
	go func() {
		b, err := io.ReadAll(reader)
		ch <- readResult{data: b, err: err}
	}()

	select {
	case <-readCtx.Done():
		return nil, readCtx.Err()
	case res := <-ch:
		return res.data, res.err
	}
}

func unknownJSONFieldMessage(err error) string {
	for current := err; current != nil; current = errors.Unwrap(current) {
		msg := current.Error()
		const token = "unknown field "
		if idx := strings.Index(msg, token); idx != -1 {
			return msg[idx:]
		}
	}
	return ""
}
