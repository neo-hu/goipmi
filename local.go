// +build linux

package goipmi

// #include <stdlib.h>
// #include "ipmi.h"
import "C"
import (
	"encoding"
	"github.com/pkg/errors"
	"sync/atomic"
	"unsafe"
)

type LocalIPMI struct {
	ctx   C.ipmi_ctx
	oem   *uint32
	close int32
}

func NewLocalIPMI() *LocalIPMI {
	return &LocalIPMI{}
}

func (l *LocalIPMI) Close() error {

	C.ipmi_close(&l.ctx)
	atomic.StoreInt32(&l.close, 1)
	return nil
}

func (l *LocalIPMI) IsClose() bool {
	return atomic.LoadInt32(&l.close) == 1
}

func (l *LocalIPMI) getSdrChunk(reservationId, recordId uint16, offset, length uint8) (uint16, []byte, error) {
	sr := &GetSdrRsp{}
	err := l.SendMessage(&GetSdrReq{
		ReservationId: reservationId,
		RecordId:      recordId,
		Offset:        offset,
		BytesToRead:   length,
	}, sr)
	if err != nil {
		return 0, nil, err
	}
	return sr.NextRecordId, sr.RecordData, nil
}

func (l *LocalIPMI) getSdrDataHelper(recordId, reservationId uint16) (uint16, []byte, error) {
	nextId, data, err := l.getSdrChunk(reservationId, recordId, 0, 5)
	if err != nil {
		return 0, nil, err
	}
	buff := NewByteBuffer(data)
	recordId, err = buff.PopUint16()
	if err != nil {
		return 0, nil, err
	}
	var recordPayloadLength uint8
	_, err = buff.PopUint8()
	if err != nil {
		return 0, nil, err
	}
	_, err = buff.PopUint8()
	if err != nil {
		return 0, nil, err
	}
	recordPayloadLength, err = buff.PopUint8()
	if err != nil {
		return 0, nil, err
	}
	recordData := data
	nextId, data, err = l.getSdrChunk(reservationId, recordId, uint8(len(data)), recordPayloadLength)
	if err != nil {
		return 0, nil, err
	}
	recordData = append(recordData, data...)
	return nextId, recordData, nil
}

func (l *LocalIPMI) GetRepositorySdr(recordId, reservationId uint16) (SdrCommon, uint16, error) {
	nextId, recordData, err := l.getSdrDataHelper(recordId, reservationId)
	if err != nil {
		return nil, nextId, err
	}
	c, err := SdrCommonFromData(recordData, nextId)
	if err != nil {
		return nil, nextId, err
	}
	return c, nextId, nil
}

func (l *LocalIPMI) getSensorReading(number uint8, ownerLun uint8) (*uint8, *int, error) {
	resp := &GetSensorReadingRsp{}
	err := l.SendMessage(&GetSensorReadingReq{
		SensorNumber: number,
		OwnerLun:     ownerLun,
	}, resp)
	if err != nil {
		return nil, nil, err
	}
	reading := &resp.SensorReading
	if resp.InitialUpdateInProgress() == 1 {
		reading = nil
	}
	var states *int
	if resp.States1 != nil {
		s1 := int(*resp.States1)
		states = &s1
		if resp.States2 != nil {
			s2 := int(*resp.States2)
			s3 := (*states) | (s2 << 8)
			states = &(s3)
		}
	}
	if resp.SensorScanningDisabled() == 0 {
		return nil, states, nil
	}
	return reading, states, nil
}

func (l *LocalIPMI) GetOem() (uint32, error) {
	if l.oem != nil {
		return *l.oem, nil
	}
	resp := &DevidRsp{}
	err := l.SendMessage(&GetOem{}, resp)
	if err != nil {
		return 0, err
	}
	oem := uint32(resp.ManufacturerId[2]&0x0F)<<16 | uint32(resp.ManufacturerId[1])<<8 | uint32(resp.ManufacturerId[0])
	l.oem = &oem
	return oem, nil
}
func (l *LocalIPMI) SelEntries(fun func([]byte) bool) error {
	resp := &GetReserveSelRsp{}
	err := l.SendMessage(&GetSelInfoReq{}, resp)
	if err != nil {
		return err
	}
	if resp.Date[0] == 0 && resp.Date[1] == 0 {
		return errors.New("SEL has no entries")
	}
	err = l.SendMessage(&ReserveSelReq{}, resp)
	if err != nil {
		return err
	}
	var nextId uint16
	var currId uint16
	var eRsp GetReserveSelRsp
	nilNextId := 2
	for nextId != uint16(0xffff) {
		currId = nextId
		if err = l.SendMessage(&GetSelEntryReq{
			Id:          currId,
			Offset:      0,
			BytesToRead: 0xff,
		}, &eRsp); err != nil {
			return err
		}
		if len(eRsp.Date) < 2 {
			return DataTooShort
		}
		nextId = uint16(eRsp.Date[0]) | uint16(eRsp.Date[1])<<8
		//
		if nextId == 0 {
			nilNextId -= 1
			if nilNextId <= 0 {
				break
			}
			continue
		}
		if !fun(eRsp.Date[2:]) {
			break
		}
		//fmt.Printf("SEL Record ID %d\n", eRsp.RecordId)
		//
		//if eRsp.RecordType == 0xf0 {
		//	fmt.Printf("Record Type           : Linux kernel panic %d\n", eRsp.RecordType)
		//	continue
		//}
		//fmt.Printf(" Record Type           : %02x", eRsp.RecordType)
		//if eRsp.RecordType >= 0xc0 {
		//	if eRsp.RecordType < 0xe0 {
		//		fmt.Print("(OEM timestamped)")
		//	} else {
		//		fmt.Print("(OEM non-timestamped)")
		//	}
		//}
		//fmt.Println("")
		//if eRsp.RecordType < 0xe0 {
		//	fmt.Print(" Timestamp             : ")
		//	if eRsp.RecordType < 0xc0 {
		//		fmt.Println(time.Unix(int64(eRsp.StandardType.Timestamp), 0))
		//	} else {
		//		fmt.Println(time.Unix(int64(eRsp.OemTsType.Timestamp), 0))
		//	}
		//}
		//if eRsp.RecordType >= 0xc0 {
		//	if eRsp.RecordType < 0xdf {
		//		fmt.Printf(" Manufactacturer ID    : %02x%02x%02x\n", eRsp.OemTsType.ManfId[0],
		//			eRsp.OemTsType.ManfId[1], eRsp.OemTsType.ManfId[2])
		//		fmt.Printf(" OEM Defined           : %v", eRsp.OemTsType.OemDefined)
		//	} else {
		//		fmt.Printf(" OEM Defined           : %v", eRsp.OemNotsType.OemDefined)
		//	}
		//	continue
		//}
		//fmt.Printf(" Generator ID          : %04x\n", eRsp.StandardType.GenId)
		//fmt.Printf(" EvM Revision          : %04x\n", eRsp.StandardType.EventType)
		//fmt.Printf(" Sensor Type           : %s\n", getSensorType(eRsp.StandardType.SensorType))
		//fmt.Printf(" Sensor Numbe          : %d\n", eRsp.StandardType.SensorNum)
		//fmt.Printf(" Event Type            : %s\n", getEventType(eRsp.StandardType.EventType))
		//val, ok := EventDirVals[eRsp.StandardType.EventDir]
		//if ok {
		//	fmt.Printf(" Event Direction       : %s\n", val)
		//} else {
		//	fmt.Printf(" Event Direction       : Unknown (0x%02X)\n", eRsp.StandardType.EventDir)
		//}
		//fmt.Printf(" Event Data (RAW)      : %02x%02x%02x\n", eRsp.StandardType.EventData[0], eRsp.StandardType.EventData[1], eRsp.StandardType.EventData[2])
		//description, err := l.GetEventDesc(eRsp)
		//if err != nil {
		//	return err
		//}
		//fmt.Printf(" Description       : %s\n", description)
	}
	return nil
}

func (l *LocalIPMI) SdrRepositoryEntries(itemFun func(string, *float64, uint8, string, uint8, uint8, string, error)) error {
	reservationId, err := l.GetReserveSDRRepoForReserveId()
	if err != nil {
		return err
	}
	recordId := uint16(0)
	for {
		c, nextId, err := l.GetRepositorySdr(recordId, reservationId)
		if recordId == uint16(0xffff) {
			break
		}

		if err != nil {
			if IsUnsupportedSDRTypeErr(err) {
				recordId = nextId
				continue
			}
			return err
		}
		recordId = nextId

		switch t := c.(type) {
		case *SdrCompactSensorRecord:
			continue
		case *SdrFullSensorRecord:
			value, _, err := l.getSensorReading(t.number, t.ownerLun)
			if err != nil {
				itemFun(t.Id, nil, 0, "", 0, 0, "", err)
				continue
			}
			if value != nil {
				val, err := t.ConvertSensorRawToValue(int(*value))
				if err != nil {
					itemFun(t.Id, nil, 0, "", 0, 0, "", err)
					continue
				}
				itemFun(t.Id, &val, t.UnitCode(), t.Unit(), t.SensorTypeCode(), t.entityInstance, t.SensorType(), nil)
			} else {
				itemFun(t.Id, nil, 0, "", 0, 0, "", nil)
			}
		default:
			continue
		}
	}
	return nil
}

func (l *LocalIPMI) GetReserveSDRRepoForReserveId() (uint16, error) {
	res := &ReserveSdrRepositoryRsp{}
	err := l.SendMessage(&ReserveSdrRepositoryReq{}, res)
	return res.ReservationId, err
}

func (l *LocalIPMI) Open() error {
	rv := C.ipmi_open(&l.ctx)
	if rv == 0 {
		return nil
	}
	return errors.Errorf("Failed to open local ipmi driver, errno is %d", rv)
}

type NetworkFunction uint8

func (l *LocalIPMI) SendMessage(req Message, resp encoding.BinaryUnmarshaler) error {
	var request C.ipmi_rq
	var response C.ipmi_rsp
	request.netfn = C.uchar(req.NetFn())
	request.lun = C.uchar(req.Lun())
	request.cmd = C.uchar(req.CmdId())
	data, err := req.MarshalBinary()
	if err != nil {
		return nil
	}
	if data != nil {
		rData := C.CBytes(data)
		defer C.free(rData)
		request.data = (*C.uchar)(rData)
		request.data_len = C.ushort(len(data))
	}
	if l.IsClose() {
		return errors.New("ipmi is close")
	}
	rv := C.ipmi_send(&l.ctx, &request, &response)
	if rv != 0 {
		return errors.Errorf("Faild to write command and recv from local ipmi driver, errno is %d", rv)
	}
	respData := C.GoBytes(unsafe.Pointer(&response.data), response.data_len)
	if CompletionCode(respData[0]) != CommandCompleted {
		return CompletionCode(respData[0])
	}
	return resp.UnmarshalBinary(respData[1:])
}
