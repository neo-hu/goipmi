// +build linux

package goipmi

import (
	"bytes"
	"encoding"
	"encoding/binary"
	"fmt"
	"github.com/pkg/errors"
	"io"
)

var (
	NetworkFunctionChassis     = NetworkFunction(0x00)
	NetworkFunctionSensorEvent = NetworkFunction(0x04)
	NetworkFunctionApp         = NetworkFunction(0x06)
	NetworkFunctionStorge      = NetworkFunction(0x0A)
	NetworkFunctionTransport   = NetworkFunction(0x0C)
)

type Command uint8

const (
	CommandGetSDRRepositoryInfo = Command(0x20)
	CommandGetReserveSDRRepo    = Command(0x22)
	CommandGetSDR               = Command(0x23)
	CommandGetSensorReading     = Command(0x2d)
)

// Command Number Assignments (table G-1)
const (
	CommandGetDeviceID              = Command(0x01)
	CommandGetAuthCapabilities      = Command(0x38)
	CommandGetSessionChallenge      = Command(0x39)
	CommandActivateSession          = Command(0x3a)
	CommandSetSessionPrivilegeLevel = Command(0x3b)
	CommandCloseSession             = Command(0x3c)
	CommandChassisControl           = Command(0x02)
	CommandChassisStatus            = Command(0x01)
	CommandSetSystemBootOptions     = Command(0x08)
	CommandGetSystemBootOptions     = Command(0x09)
	// CommandGetSDRRepositoryInfo     = Command(0x20)

	// CommandGetReserveSDRRepo     	= Command(0x22)
)

type Request struct {
	NetworkFunction
	Command
	Data interface{}
}

type Response interface {
	Code() uint8
}

func messageDataToBytes(data interface{}) []byte {
	if encoder, ok := data.(encoding.BinaryMarshaler); ok {
		buf, err := encoder.MarshalBinary()
		if err != nil {
			panic(err)
		}
		return buf
	}
	buf := new(bytes.Buffer)
	binaryWrite(buf, data)
	return buf.Bytes()
}

func binaryWrite(writer io.Writer, data interface{}) {
	err := binary.Write(writer, binary.LittleEndian, data)
	if err != nil {
		// shouldn't happen to a bytes.Buffer
		panic(err)
	}
}

type Message interface {
	NetFn() NetworkFunction
	CmdId() Command
	Lun() uint8
	encoding.BinaryMarshaler
}

type GetSelEntryReq struct {
	Id          uint16
	Offset      uint8
	BytesToRead uint8
}

func (r *GetSelEntryReq) String() string {
	return fmt.Sprintf("<GetSdrReq Id=%d, Offset=%d, BytesToRead=%d>", r.Id, r.Offset, r.BytesToRead)
}
func (r *GetSelEntryReq) Lun() uint8 {
	return 0
}

func (r *GetSelEntryReq) NetFn() NetworkFunction {
	return NetworkFunctionStorge
}
func (r *GetSelEntryReq) CmdId() Command {
	return 0x43
}

func (r *GetSelEntryReq) MarshalBinary() ([]byte, error) {
	data := make([]byte, 6)
	//binary.LittleEndian.PutUint16(data, r.ReservationId)
	// todo ReservationId
	data[0] = 0x00
	data[1] = 0x00

	// todo RecordId
	data[2] = byte(r.Id)
	data[3] = byte(r.Id >> 8)

	data[4] = r.Offset
	data[5] = r.BytesToRead
	return data, nil
}

type GetSelInfoReq struct {
}

func (r *GetSelInfoReq) MarshalBinary() (data []byte, err error) {
	return nil, nil
}

func (r *GetSelInfoReq) String() string {
	return "<GetSelInfoReq>"
}
func (r *GetSelInfoReq) Lun() uint8 {
	return 0
}

func (r *GetSelInfoReq) NetFn() NetworkFunction {
	return NetworkFunctionStorge
}
func (r *GetSelInfoReq) CmdId() Command {
	return 0x40
}

type DevidRsp struct {
	DeviceId          uint8
	DeviceRevision    uint8
	FwRev1            uint8
	FwRev2            uint8
	IpmiVersion       uint8
	AdtlDeviceSupport uint8
	ManufacturerId    [3]uint8
	ProductId         [2]uint8
	AuxFwRev          [4]uint8
}

func (r *DevidRsp) String() string {
	return fmt.Sprintf("DevidRsp DeviceId=%v ManufacturerId=%v", r.DeviceId, r.ManufacturerId)
}
func (r *DevidRsp) UnmarshalBinary(data []byte) (err error) {
	buff := NewByteBuffer(data)
	if r.DeviceId, err = buff.PopUint8(); err != nil {
		return err
	}
	if r.DeviceRevision, err = buff.PopUint8(); err != nil {
		return err
	}
	if r.FwRev1, err = buff.PopUint8(); err != nil {
		return err
	}
	if r.FwRev2, err = buff.PopUint8(); err != nil {
		return err
	}
	if r.IpmiVersion, err = buff.PopUint8(); err != nil {
		return err
	}
	if r.AdtlDeviceSupport, err = buff.PopUint8(); err != nil {
		return err
	}
	if r.ManufacturerId[0], err = buff.PopUint8(); err != nil {
		return err
	}
	if r.ManufacturerId[1], err = buff.PopUint8(); err != nil {
		return err
	}
	if r.ManufacturerId[2], err = buff.PopUint8(); err != nil {
		return err
	}
	if r.ProductId[0], err = buff.PopUint8(); err != nil {
		return err
	}
	if r.ProductId[1], err = buff.PopUint8(); err != nil {
		return err
	}
	if r.AuxFwRev[0], err = buff.PopUint8(); err != nil {
		return err
	}
	if r.AuxFwRev[1], err = buff.PopUint8(); err != nil {
		return err
	}
	if r.AuxFwRev[2], err = buff.PopUint8(); err != nil {
		return err
	}
	if r.AuxFwRev[3], err = buff.PopUint8(); err != nil {
		return err
	}
	return nil
}

type GetReserveSelRsp struct {
	Date []byte
}

func (r *GetReserveSelRsp) String() string {
	return fmt.Sprintf("GetReserveSelRsp Data=%v", r.Date)
}
func (r *GetReserveSelRsp) UnmarshalBinary(data []byte) error {
	r.Date = data
	return nil
}

type GetOem struct {
}

func (r *GetOem) MarshalBinary() (data []byte, err error) {
	return nil, nil
}

func (r *GetOem) String() string {
	return "<GetOem>"
}
func (r *GetOem) Lun() uint8 {
	return 0
}

func (r *GetOem) NetFn() NetworkFunction {
	return NetworkFunctionApp
}
func (r *GetOem) CmdId() Command {
	return 0x01
}

type ReserveSelReq struct {
}

func (r *ReserveSelReq) MarshalBinary() (data []byte, err error) {
	return nil, nil
}

func (r *ReserveSelReq) String() string {
	return "<ReserveSelReq>"
}
func (r *ReserveSelReq) Lun() uint8 {
	return 0
}

func (r *ReserveSelReq) NetFn() NetworkFunction {
	return NetworkFunctionStorge
}
func (r *ReserveSelReq) CmdId() Command {
	return 0x42
}

type ReserveSdrRepositoryReq struct {
}

func (r *ReserveSdrRepositoryReq) MarshalBinary() (data []byte, err error) {
	return nil, nil
}

func (r *ReserveSdrRepositoryReq) String() string {
	return "<ReserveSdrRepositoryReq>"
}
func (r *ReserveSdrRepositoryReq) Lun() uint8 {
	return 0
}

func (r *ReserveSdrRepositoryReq) NetFn() NetworkFunction {
	return NetworkFunctionStorge
}
func (r *ReserveSdrRepositoryReq) CmdId() Command {
	return CommandGetReserveSDRRepo
}

type ReserveSdrRepositoryRsp struct {
	ReservationId uint16
}

func (r *ReserveSdrRepositoryRsp) String() string {
	return fmt.Sprintf("<ReserveSdrRepositoryRsp ReservationId=%d>", r.ReservationId)
}
func (r *ReserveSdrRepositoryRsp) UnmarshalBinary(data []byte) error {
	if len(data) < 2 {
		return errors.Errorf("invalid data len:%d < 2", len(data))
	}
	r.ReservationId = uint16(data[0]) | uint16(data[1])<<8
	return nil
}

type GetSdrReq struct {
	ReservationId uint16
	RecordId      uint16
	Offset        uint8
	BytesToRead   uint8
}

func (r *GetSdrReq) SetReservationId(reservationId uint16) {
	r.ReservationId = reservationId
}

func (r *GetSdrReq) String() string {
	return fmt.Sprintf("<GetSdrReq ReservationId=%d, RecordId=%d, Offset=%d, BytesToRead=%d>", r.ReservationId, r.RecordId, r.Offset, r.BytesToRead)
}
func (r *GetSdrReq) Lun() uint8 {
	return 0
}

func (r *GetSdrReq) NetFn() NetworkFunction {
	return NetworkFunctionStorge
}
func (r *GetSdrReq) CmdId() Command {
	return CommandGetSDR
}

func (r *GetSdrReq) MarshalBinary() ([]byte, error) {
	data := make([]byte, 6)
	//binary.LittleEndian.PutUint16(data, r.ReservationId)
	// todo ReservationId
	data[0] = byte(r.ReservationId)
	data[1] = byte(r.ReservationId >> 8)

	// todo RecordId
	data[2] = byte(r.RecordId)
	data[3] = byte(r.RecordId >> 8)

	data[4] = r.Offset
	data[5] = r.BytesToRead
	return data, nil
}

type GetSdrRsp struct {
	NextRecordId uint16
	RecordData   []byte
}

func (r *GetSdrRsp) String() string {
	return fmt.Sprintf("<GetSdrRsp NextRecordId=%d, RecordData=%d>", r.NextRecordId, len(r.RecordData))
}
func (r *GetSdrRsp) UnmarshalBinary(data []byte) error {
	if len(data) < 2 {
		return errors.Errorf("invalid data len:%d < 2", len(data))
	}
	r.NextRecordId = uint16(data[0]) | uint16(data[1])<<8
	r.RecordData = data[2:]
	return nil
}

type GetSensorReadingRsp struct {
	SensorReading uint8
	Config        uint8
	States1       *uint8
	States2       *uint8
}

func (r *GetSensorReadingRsp) String() string {
	return fmt.Sprintf("<GetSensorReadingRsp SensorReading=%d, Config=%d, States1=%d, States2=%d>", r.SensorReading, r.Config, r.States1, r.States2)
}

func (r *GetSensorReadingRsp) InitialUpdateInProgress() uint8 {
	return (r.Config >> 5) & 1
}

func (r *GetSensorReadingRsp) SensorScanningDisabled() uint8 {
	return (r.Config >> 6) & 1
}
func (r *GetSensorReadingRsp) EventMessageDisabled() uint8 {
	return (r.Config >> 7) & 1
}
func (r *GetSensorReadingRsp) UnmarshalBinary(data []byte) error {
	if len(data) < 2 {
		return errors.Errorf("invalid data len:%d  < 2", len(data))
	}
	r.SensorReading = data[0]
	r.Config = data[1]
	if len(data) > 2 {
		r.States1 = &data[2]
	}
	if len(data) > 3 {
		r.States2 = &data[3]
	}
	return nil
}

type GetSensorReadingReq struct {
	SensorNumber uint8
	OwnerLun     uint8
}

func (r *GetSensorReadingReq) MarshalBinary() ([]byte, error) {
	data := make([]byte, 1)
	data[0] = r.SensorNumber
	return data, nil
}

func (r *GetSensorReadingReq) String() string {
	return fmt.Sprintf("<GetSensorReadingReq SensorNumber=%d>", r.SensorNumber)
}
func (r *GetSensorReadingReq) Lun() uint8 {
	if r.OwnerLun != 0 {
		return r.OwnerLun
	}
	return 0
}

func (r *GetSensorReadingReq) NetFn() NetworkFunction {
	return NetworkFunctionSensorEvent
}
func (r *GetSensorReadingReq) CmdId() Command {
	return CommandGetSensorReading
}
