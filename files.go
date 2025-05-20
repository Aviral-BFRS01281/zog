package zog

import (
	"mime/multipart"

	"github.com/Aviral-BFRS01281/zog/conf"
	p "github.com/Aviral-BFRS01281/zog/internals"
	"github.com/Aviral-BFRS01281/zog/zconst"
)

var _ PrimitiveZogSchema[multipart.FileHeader] = &FileSchema[multipart.FileHeader]{}

type FileSchema[T multipart.FileHeader] struct {
	processors []p.ZProcessor[*T]
	defaultVal *T
	required   *p.Test[*T]
	catch      *T
	coercer    CoercerFunc
}

// ! INTERNALS

// Returns the type of the schema
func (v *FileSchema[T]) getType() zconst.ZogType {
	return zconst.TypeFile
}

// Sets the coercer for the schema
func (v *FileSchema[T]) setCoercer(c CoercerFunc) {
	v.coercer = c
}

// Returns a new String Shape
func File(opts ...SchemaOption) *FileSchema[multipart.FileHeader] {
	s := &FileSchema[multipart.FileHeader]{
		coercer: conf.Coercers.File, // default coercer
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// parses the value and stores it in the destination
func (v *FileSchema[T]) Parse(data any, dest *T, options ...ExecOption) ZogIssueList {
	errs := p.NewErrsList()
	defer errs.Free()
	ctx := p.NewExecCtx(errs, conf.IssueFormatter)
	defer ctx.Free()
	for _, opt := range options {
		opt(ctx)
	}

	path := p.NewPathBuilder()
	defer path.Free()
	sctx := ctx.NewSchemaCtx(data, dest, path, v.getType())
	defer sctx.Free()
	v.process(sctx)
	return errs.List
}

// Internal function to process the data
func (v *FileSchema[T]) process(ctx *p.SchemaCtx) {
	primitiveParsing(ctx, v.processors, v.defaultVal, v.required, v.catch, v.coercer, p.IsParseZeroValue)
}

// Validates a number pointer
func (v *FileSchema[T]) Validate(data *T, options ...ExecOption) ZogIssueList {
	errs := p.NewErrsList()
	defer errs.Free()
	ctx := p.NewExecCtx(errs, conf.IssueFormatter)
	defer ctx.Free()
	for _, opt := range options {
		opt(ctx)
	}

	path := p.NewPathBuilder()
	defer path.Free()
	sctx := ctx.NewSchemaCtx(data, data, path, v.getType())
	defer sctx.Free()
	v.validate(sctx)
	return errs.List
}

func (v *FileSchema[T]) validate(ctx *p.SchemaCtx) {
	primitiveValidation(ctx, v.processors, v.defaultVal, v.required, v.catch)
}

// marks field as required
func (v *FileSchema[T]) Required(options ...TestOption) *FileSchema[T] {
	r := p.Required[*T]()
	for _, opt := range options {
		opt(&r)
	}
	v.required = &r
	return v
}

// marks field as optional
func (v *FileSchema[T]) Optional() *FileSchema[T] {
	v.required = nil
	return v
}

// sets the default value
func (v *FileSchema[T]) Default(val T) *FileSchema[T] {
	v.defaultVal = &val
	return v
}

// sets the default value
func (v *FileSchema[T]) Default2(val T) *FileSchema[T] {
	v.defaultVal = &val
	return v
}
