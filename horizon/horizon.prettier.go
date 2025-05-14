package horizon

import (
	"bytes"
	"encoding/json"
	"time"

	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

type HorizonPrettyJSONEncoder struct {
	encoder zapcore.Encoder
}

func NewHorizonPrettyJSONEncoder(cfg zapcore.EncoderConfig) zapcore.Encoder {
	enc := zapcore.NewJSONEncoder(cfg)
	return &HorizonPrettyJSONEncoder{encoder: enc}
}

func (p *HorizonPrettyJSONEncoder) Clone() zapcore.Encoder {
	return &HorizonPrettyJSONEncoder{encoder: p.encoder.Clone()}
}

func (p *HorizonPrettyJSONEncoder) EncodeEntry(ent zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	buf, err := p.encoder.EncodeEntry(ent, fields)
	if err != nil {
		return nil, err
	}

	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, buf.Bytes(), "", "  ")
	if err != nil {
		return buf, nil
	}

	newBuf := buffer.NewPool().Get()
	newBuf.Write(prettyJSON.Bytes())
	newBuf.WriteByte('\n')

	return newBuf, nil
}

// Forward other required zapcore.Encoder methods to underlying encoder

func (p *HorizonPrettyJSONEncoder) AddArray(key string, marshaler zapcore.ArrayMarshaler) error {
	return p.encoder.AddArray(key, marshaler)
}

func (p *HorizonPrettyJSONEncoder) AddObject(key string, marshaler zapcore.ObjectMarshaler) error {
	return p.encoder.AddObject(key, marshaler)
}

func (p *HorizonPrettyJSONEncoder) AddBinary(key string, value []byte) {
	p.encoder.AddBinary(key, value)
}

func (p *HorizonPrettyJSONEncoder) AddByteString(key string, value []byte) {
	p.encoder.AddByteString(key, value)
}

func (p *HorizonPrettyJSONEncoder) AddBool(key string, value bool) {
	p.encoder.AddBool(key, value)
}

func (p *HorizonPrettyJSONEncoder) AddComplex128(key string, value complex128) {
	p.encoder.AddComplex128(key, value)
}

func (p *HorizonPrettyJSONEncoder) AddComplex64(key string, value complex64) {
	p.encoder.AddComplex64(key, value)
}

func (p *HorizonPrettyJSONEncoder) AddDuration(key string, value time.Duration) {
	p.encoder.AddDuration(key, value)
}

func (p *HorizonPrettyJSONEncoder) AddFloat64(key string, value float64) {
	p.encoder.AddFloat64(key, value)
}

func (p *HorizonPrettyJSONEncoder) AddFloat32(key string, value float32) {
	p.encoder.AddFloat32(key, value)
}

func (p *HorizonPrettyJSONEncoder) AddInt(key string, value int) {
	p.encoder.AddInt(key, value)
}

func (p *HorizonPrettyJSONEncoder) AddInt64(key string, value int64) {
	p.encoder.AddInt64(key, value)
}

func (p *HorizonPrettyJSONEncoder) AddInt32(key string, value int32) {
	p.encoder.AddInt32(key, value)
}

func (p *HorizonPrettyJSONEncoder) AddInt16(key string, value int16) {
	p.encoder.AddInt16(key, value)
}

func (p *HorizonPrettyJSONEncoder) AddInt8(key string, value int8) {
	p.encoder.AddInt8(key, value)
}

func (p *HorizonPrettyJSONEncoder) AddString(key string, value string) {
	p.encoder.AddString(key, value)
}

func (p *HorizonPrettyJSONEncoder) AddTime(key string, value time.Time) {
	p.encoder.AddTime(key, value)
}

func (p *HorizonPrettyJSONEncoder) AddUint(key string, value uint) {
	p.encoder.AddUint(key, value)
}

func (p *HorizonPrettyJSONEncoder) AddUint64(key string, value uint64) {
	p.encoder.AddUint64(key, value)
}

func (p *HorizonPrettyJSONEncoder) AddUint32(key string, value uint32) {
	p.encoder.AddUint32(key, value)
}

func (p *HorizonPrettyJSONEncoder) AddUint16(key string, value uint16) {
	p.encoder.AddUint16(key, value)
}

func (p *HorizonPrettyJSONEncoder) AddUint8(key string, value uint8) {
	p.encoder.AddUint8(key, value)
}

func (p *HorizonPrettyJSONEncoder) AddUintptr(key string, value uintptr) {
	p.encoder.AddUintptr(key, value)
}

func (p *HorizonPrettyJSONEncoder) AddReflected(key string, value any) error {
	return p.encoder.AddReflected(key, value)
}

func (p *HorizonPrettyJSONEncoder) OpenNamespace(key string) {
	p.encoder.OpenNamespace(key)
}
