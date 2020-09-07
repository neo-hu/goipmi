// +build linux

package goipmi

import (
	"github.com/pkg/errors"
	"math"
)

var (
	SDR_TYPE_FULL_SENSOR_RECORD                          = byte(0x01)
	SDR_TYPE_COMPACT_SENSOR_RECORD                       = byte(0x02)
	SDR_TYPE_EVENT_ONLY_SENSOR_RECORD                    = byte(0x03)
	SDR_TYPE_ENTITY_ASSOCIATION_RECORD                   = byte(0x08)
	SDR_TYPE_FRU_DEVICE_LOCATOR_RECORD                   = byte(0x11)
	SDR_TYPE_MANAGEMENT_CONTROLLER_DEVICE_LOCATOR_RECORD = byte(0x12)
	SDR_TYPE_MANAGEMENT_CONTROLLER_CONFIRMATION_RECORD   = byte(0x13)
	SDR_TYPE_BMC_MESSAGE_CHANNEL_INFO_RECORD             = byte(0x14)
	SDR_TYPE_OEM_SENSOR_RECORD                           = byte(0xC0)
)
var sdrRecordValueSensorType []string = []string{
	"reserved",
	"Temperature", "Voltage", "Current", "Fan",
	"Physical Security", "Platform Security", "Processor",
	"Power Supply", "Power Unit", "Cooling Device", "Other",
	"Memory", "Drive Slot / Bay", "POST Memory Resize",
	"System Firmwares", "Event Logging Disabled", "Watchdog1",
	"System Event", "Critical Interrupt", "Button",
	"Module / Board", "Microcontroller", "Add-in Card",
	"Chassis", "Chip Set", "Other FRU", "Cable / Interconnect",
	"Terminator", "System Boot Initiated", "Boot Error",
	"OS Boot", "OS Critical Stop", "Slot / Connector",
	"System ACPI Power State", "Watchdog2", "Platform Alert",
	"Entity Presence", "Monitor ASIC", "LAN",
	"Management Subsys Health", "Battery", "Session Audit",
	"Version Change", "FRU State"}
var sdrRecordValueBasicUnit []string = []string{
	"unspecified",
	"degrees C", "degrees F", "degrees K",
	"Volts", "Amps", "Watts", "Joules",
	"Coulombs", "VA", "Nits",
	"lumen", "lux", "Candela",
	"kPa", "PSI", "Newton",
	"CFM", "RPM", "Hz",
	"microsecond", "millisecond", "second", "minute", "hour",
	"day", "week", "mil", "inches", "feet", "cu in", "cu feet",
	"mm", "cm", "m", "cu cm", "cu m", "liters", "fluid ounce",
	"radians", "steradians", "revolutions", "cycles",
	"gravities", "ounce", "pound", "ft-lb", "oz-in", "gauss",
	"gilberts", "henry", "millihenry", "farad", "microfarad",
	"ohms", "siemens", "mole", "becquerel", "PPM", "reserved",
	"Decibels", "DbA", "DbC", "gray", "sievert",
	"color temp deg K", "bit", "kilobit", "megabit", "gigabit",
	"byte", "kilobyte", "megabyte", "gigabyte", "word", "dword",
	"qword", "line", "hit", "miss", "retry", "reset",
	"overflow", "underrun", "collision", "packets", "messages",
	"characters", "error", "correctable error", "uncorrectable error"}

type SdrCommon interface {
}

func SdrCommonFromData(recordData []byte, nextId uint16) (SdrCommon, error) {
	if len(recordData) < 4 {
		return nil, DataTooShort
	}
	sdrType := recordData[3]
	switch sdrType {
	case SDR_TYPE_COMPACT_SENSOR_RECORD:
		return NewSdrCompactSensorRecord(recordData, nextId)
	case SDR_TYPE_FULL_SENSOR_RECORD:
		return NewSdrFullSensorRecord(recordData, nextId)
	default:
		return nil, NewUnsupportedSDRTypeErr(sdrType, nextId)
	}
}

type SdrCompactSensorRecord struct {
	SdrCommonHeader
	nextId                    uint16
	Data                      []byte
	ownerId, ownerLun, number uint8
	entityId, entityInstance  uint8
	Id                        string
	reserved                  uint32
	oem                       uint8
	capabilities              uint8
	initialization            uint8
	sensorTypeCode            uint8
	readingType               uint8
	assertionMask             uint16
	deassertionMask           uint16
	discreteReadingMask       uint16
	recordSharing             uint16
	positiveGoingHysteresis   uint8
	negativeGoingHysteresis   uint8
	units1, units2, units3    uint8
}

func NewSdrCompactSensorRecord(data []byte, nextId uint16) (SdrCommon, error) {
	header, err := CommonHeader(data)
	if err != nil {
		return nil, err
	}
	s := &SdrCompactSensorRecord{SdrCommonHeader: header, Data: data, nextId: nextId}
	return s, s.UnmarshalBinary(data)
}

func (s *SdrCompactSensorRecord) UnmarshalBinary(data []byte) error {
	buff := NewByteBuffer(data[5:])
	var err error
	// record key bytes
	s.ownerId, err = buff.PopUint8() // 6

	if err != nil {
		return err
	}
	s.ownerLun, err = buff.PopUint8() // 7
	if err != nil {
		return err
	}
	s.ownerLun = s.ownerLun & 0x3
	s.number, err = buff.PopUint8() // 8
	if err != nil {
		return err
	}

	// record body bytes
	s.entityId, err = buff.PopUint8() // 9
	if err != nil {
		return err
	}
	s.entityInstance, err = buff.PopUint8() // 10
	if err != nil {
		return err
	}

	s.initialization, err = buff.PopUint8() // 11
	if err != nil {
		return err
	}
	s.capabilities, err = buff.PopUint8() // 12
	if err != nil {
		return err
	}
	s.sensorTypeCode, err = buff.PopUint8() // 13 SensorType
	s.readingType, err = buff.PopUint8()    // 14 ReadingType
	if err != nil {
		return err
	}
	s.assertionMask, err = buff.PopUint16() // 16 AssertionEventMask
	if err != nil {
		return err
	}
	s.deassertionMask, err = buff.PopUint16() // 18 DeassertionEventMask
	if err != nil {
		return err
	}
	s.discreteReadingMask, err = buff.PopUint16() // 20 DiscreteReadingMask
	if err != nil {
		return err
	}

	s.units1, err = buff.PopUint8() // 21 Unit
	if err != nil {
		return err
	}

	s.units2, err = buff.PopUint8() // 22 BaseUnit
	if err != nil {
		return err
	}
	s.units3, err = buff.PopUint8() // 23 ModifierUnit
	if err != nil {
		return err
	}
	s.recordSharing, err = buff.PopUint16() // 25 SensorRecSharing
	if err != nil {
		return err
	}
	s.positiveGoingHysteresis, err = buff.PopUint8() // 26 PThresHysteresisVal
	if err != nil {
		return err
	}
	s.negativeGoingHysteresis, err = buff.PopUint8() // 27 NThresHysteresisVal
	if err != nil {
		return err
	}
	s.reserved, err = buff.PopUint24() // 31 Reserved
	if err != nil {
		return err
	}
	s.oem, err = buff.PopUint8() // 32
	if err != nil {
		return err
	}
	s.Id, err = deviceIdString(buff)
	return err
}

type SdrFullSensorRecord struct {
	Id string
	SdrCommonHeader
	ownerId, ownerLun, number                          uint8
	entityId, entityInstance                           uint8
	nextId                                             uint16
	Data                                               []byte
	initialization                                     []string
	capabilities                                       []string
	analogCharacteristic                               []string
	sensorTypeCode                                     uint8
	eventReadingTypeCode                               uint8
	assertionMask                                      uint16
	deassertionMask                                    uint16
	discreteReadingMask                                uint16
	units1, units2, units3                             uint8
	analogDataFormat, rateUnit                         uint8
	modifierUnit, percentage, linearization            uint8
	tolerance                                          uint8
	m                                                  int
	b                                                  int
	accuracy                                           int
	accuracyExp                                        int
	k2, k1                                             int
	sensorMinimumReading, nominalReading               uint8
	normalMaximum, normalMinimum, sensorMaximumReading uint8

	threshold  map[string]uint8
	hysteresis map[string]uint8
	reserved   uint16
	oem        uint8
}

func NewSdrFullSensorRecord(data []byte, nextId uint16) (SdrCommon, error) {
	header, err := CommonHeader(data)
	if err != nil {
		return nil, err
	}
	s := &SdrFullSensorRecord{SdrCommonHeader: header, Data: data, nextId: nextId}
	return s, s.UnmarshalBinary(data)
}

func (s *SdrFullSensorRecord) UnmarshalBinary(data []byte) error {
	buff := NewByteBuffer(data[5:])
	var err error
	// record key bytes
	s.ownerId, err = buff.PopUint8()
	if err != nil {
		return err
	}
	s.ownerLun, err = buff.PopUint8()
	if err != nil {
		return err
	}
	s.ownerLun = s.ownerLun & 0x3
	s.number, err = buff.PopUint8()
	if err != nil {
		return err
	}
	// record body bytes
	s.entityId, err = buff.PopUint8()
	if err != nil {
		return err
	}
	s.entityInstance, err = buff.PopUint8()
	if err != nil {
		return err
	}
	initialization, err := buff.PopUint8()
	if err != nil {
		return err
	}
	s.initialization = make([]string, 0, 7)
	if initialization&0x40 != 0 {
		s.initialization = append(s.initialization, "scanning")
	}
	if initialization&0x20 != 0 {
		s.initialization = append(s.initialization, "events")
	}
	if initialization&0x10 != 0 {
		s.initialization = append(s.initialization, "thresholds")
	}
	if initialization&0x08 != 0 {
		s.initialization = append(s.initialization, "hysteresis")
	}
	if initialization&0x04 != 0 {
		s.initialization = append(s.initialization, "type")
	}
	if initialization&0x02 != 0 {
		s.initialization = append(s.initialization, "default_event_generation")
	}
	if initialization&0x01 != 0 {
		s.initialization = append(s.initialization, "default_scanning")
	}

	c, err := buff.PopUint8()
	if err != nil {
		return err
	}
	s.decodeCapabilities(int(c))

	s.sensorTypeCode, err = buff.PopUint8()
	if err != nil {
		return err
	}
	s.eventReadingTypeCode, err = buff.PopUint8()
	if err != nil {
		return err
	}
	s.assertionMask, err = buff.PopUint16()
	if err != nil {
		return err
	}
	s.deassertionMask, err = buff.PopUint16()
	if err != nil {
		return err
	}
	s.discreteReadingMask, err = buff.PopUint16()
	if err != nil {
		return err
	}

	s.units1, err = buff.PopUint8()
	if err != nil {
		return err
	}

	s.units2, err = buff.PopUint8()
	if err != nil {
		return err
	}
	s.units3, err = buff.PopUint8()
	if err != nil {
		return err
	}
	s.analogDataFormat = (s.units1 >> 6) & 0x3
	s.rateUnit = (s.units1 >> 6) >> 0x7
	s.modifierUnit = (s.units1 >> 1) & 0x2
	s.percentage = s.units1 & 0x1

	s.linearization, err = buff.PopUint8()
	if err != nil {
		return err
	}
	s.linearization = s.linearization & 0x7f

	//
	m, err := buff.PopUint8()
	if err != nil {
		return err
	}
	mTol, err := buff.PopUint8()
	if err != nil {
		return err
	}
	s.m = (int(m) & 0xff) | ((int(mTol) & 0xc0) << 2)
	s.m = ConvertComplement(s.m, 10)
	s.tolerance = mTol & 0x3f

	b, err := buff.PopUint8()
	if err != nil {
		return err
	}

	bAcc, err := buff.PopUint8()
	if err != nil {
		return err
	}
	s.b = (int(b) & 0xff) | ((int(bAcc) & 0xc0) << 2)
	s.b = ConvertComplement(s.b, 10)

	accAccexp, err := buff.PopUint8()
	if err != nil {
		return err
	}
	s.accuracy = (int(bAcc) & 0x3f) | ((int(accAccexp) & 0xf0) << 4)
	s.accuracyExp = (int(accAccexp) & 0x0c) >> 2

	rexpBexp, err := buff.PopUint8()
	if err != nil {
		return err
	}
	s.k2 = (int(rexpBexp) & 0xf0) >> 4
	s.k2 = ConvertComplement(s.k2, 4)

	s.k1 = int(rexpBexp & 0x0f)
	s.k1 = ConvertComplement(s.k1, 4)

	// 31
	analogCharacteristics, err := buff.PopUint8()
	if err != nil {
		return err
	}
	s.analogCharacteristic = make([]string, 0, 3)
	if analogCharacteristics&0x01 != 0 {
		s.analogCharacteristic = append(s.analogCharacteristic, "nominal_reading")
	}
	if analogCharacteristics&0x02 != 0 {
		s.analogCharacteristic = append(s.analogCharacteristic, "normal_max")
	}
	if analogCharacteristics&0x04 != 0 {
		s.analogCharacteristic = append(s.analogCharacteristic, "normal_min")
	}
	s.nominalReading, err = buff.PopUint8()
	if err != nil {
		return err
	}
	s.normalMaximum, err = buff.PopUint8()
	if err != nil {
		return err
	}
	s.normalMinimum, err = buff.PopUint8()
	if err != nil {
		return err
	}
	s.sensorMaximumReading, err = buff.PopUint8()
	if err != nil {
		return err
	}
	s.sensorMinimumReading, err = buff.PopUint8()
	if err != nil {
		return err
	}

	s.threshold = map[string]uint8{}
	s.threshold["unr"], err = buff.PopUint8()
	if err != nil {
		return err
	}
	s.threshold["ucr"], err = buff.PopUint8()
	if err != nil {
		return err
	}
	s.threshold["unc"], err = buff.PopUint8()
	if err != nil {
		return err
	}
	s.threshold["lnr"], err = buff.PopUint8()
	if err != nil {
		return err
	}
	s.threshold["lcr"], err = buff.PopUint8()
	if err != nil {
		return err
	}
	s.threshold["lnc"], err = buff.PopUint8()
	if err != nil {
		return err
	}

	s.hysteresis = map[string]uint8{}
	s.hysteresis["positive_going"], err = buff.PopUint8()
	if err != nil {
		return err
	}
	s.hysteresis["negative_going"], err = buff.PopUint8()
	if err != nil {
		return err
	}

	s.reserved, err = buff.PopUint16()
	if err != nil {
		return err
	}
	s.oem, err = buff.PopUint8()
	if err != nil {
		return err
	}
	s.Id, err = deviceIdString(buff)
	if err != nil {
		return err
	}
	return nil
}
func (s *SdrFullSensorRecord) UnitCode() uint8 {
	return s.units2
}
func (s *SdrFullSensorRecord) Unit() string {
	if s.units2 >= 0 && s.units2 < uint8(len(sdrRecordValueBasicUnit)) {
		return sdrRecordValueBasicUnit[s.units2]
	}
	return ""
}

func (s *SdrFullSensorRecord) SensorTypeCode() uint8 {
	return s.sensorTypeCode
}
func (s *SdrFullSensorRecord) SensorType() string {
	if s.sensorTypeCode >= 0 && s.sensorTypeCode < uint8(len(sdrRecordValueSensorType)) {
		return sdrRecordValueSensorType[s.sensorTypeCode]
	}
	return ""
}
func (s *SdrFullSensorRecord) decodeCapabilities(capabilities int) {
	s.capabilities = make([]string, 0, 10)
	if capabilities&0x80 != 0 {
		s.capabilities = append(s.capabilities, "ignore_sensor")
	}
	if capabilities&0x40 != 0 {
		s.capabilities = append(s.capabilities, "auto_rearm")
	}
	var (
		HYSTERESIS_MASK                 = 0x30
		HYSTERESIS_IS_NOT_SUPPORTED     = 0x00
		HYSTERESIS_IS_READABLE          = 0x10
		HYSTERESIS_IS_READ_AND_SETTABLE = 0x20
		HYSTERESIS_IS_FIXED             = 0x30
	)
	if capabilities&HYSTERESIS_MASK == HYSTERESIS_IS_NOT_SUPPORTED {
		s.capabilities = append(s.capabilities, "hysteresis_not_supported")
	} else if capabilities&HYSTERESIS_MASK == HYSTERESIS_IS_READABLE {
		s.capabilities = append(s.capabilities, "hysteresis_readable")
	} else if capabilities&HYSTERESIS_MASK == HYSTERESIS_IS_READ_AND_SETTABLE {
		s.capabilities = append(s.capabilities, "hysteresis_read_and_setable")
	} else if capabilities&HYSTERESIS_MASK == HYSTERESIS_IS_FIXED {
		s.capabilities = append(s.capabilities, "hysteresis_fixed")
	}

	var (
		THRESHOLD_MASK                 = 0x0C
		THRESHOLD_IS_NOT_SUPPORTED     = 0x00
		THRESHOLD_IS_READABLE          = 0x08
		THRESHOLD_IS_READ_AND_SETTABLE = 0x04
		THRESHOLD_IS_FIXED             = 0x0C
	)
	if capabilities&THRESHOLD_MASK == THRESHOLD_IS_NOT_SUPPORTED {
		s.capabilities = append(s.capabilities, "threshold_not_supported")
	} else if capabilities&THRESHOLD_MASK == THRESHOLD_IS_READABLE {
		s.capabilities = append(s.capabilities, "threshold_readable")
	} else if capabilities&THRESHOLD_MASK == THRESHOLD_IS_READ_AND_SETTABLE {
		s.capabilities = append(s.capabilities, "threshold_read_and_setable")
	} else if capabilities&THRESHOLD_MASK == THRESHOLD_IS_FIXED {
		s.capabilities = append(s.capabilities, "threshold_fixed")
	}
}

func ConvertComplement(value, size int) int {
	if value&(1<<(uint(size)-1)) != 0 {
		value = (-(1 << uint(size))) + value
	}
	return value
}

func deviceIdString(buff *ByteBuffer) (string, error) {
	length, err := buff.PopUint8()
	if err != nil {
		return "", err
	}
	length = parseIdLen(length)
	return buff.PopString(int(length))
}

func parseIdLen(len uint8) uint8 {
	return len & 0x1f
}

type SdrCommonHeader struct {
	id      uint16
	version uint8
	typ     uint8
	length  uint8
}

func CommonHeader(data []byte) (SdrCommonHeader, error) {
	buff := NewByteBuffer(data)
	recordId, err := buff.PopUint16()

	var recordVersion uint8
	var recordType uint8
	var recordPayloadLength uint8
	recordVersion, err = buff.PopUint8()
	if err != nil {
		return SdrCommonHeader{}, err
	}
	recordType, err = buff.PopUint8()
	if err != nil {
		return SdrCommonHeader{}, err
	}
	recordPayloadLength, err = buff.PopUint8()
	if err != nil {
		return SdrCommonHeader{}, err
	}

	return SdrCommonHeader{id: recordId, version: recordVersion, typ: recordType, length: recordPayloadLength}, nil

}

const (
	L_LINEAR = 0
	L_LN     = 1
	L_LOG    = 2
	L_LOG2   = 3
	L_E      = 4
	L_EXP10  = 5
	L_EXP2   = 6
	L_1_X    = 7
	L_SQR    = 8
	L_CUBE   = 9
	L_SQRT   = 10
	L_CUBERT = 11
)

// SDR type 0x01
const (
	DATA_FMT_UNSIGNED      = uint8(0)
	DATA_FMT_1S_COMPLEMENT = uint8(1)
	DATA_FMT_2S_COMPLEMENT = uint8(2)
	DATA_FMT_NONE          = uint8(3)
)

func (s *SdrFullSensorRecord) ConvertSensorRawToValue(raw int) (float64, error) {
	switch s.analogDataFormat {
	case DATA_FMT_1S_COMPLEMENT:
		if raw&0x80 != 0 {
			raw = -((raw & 0x7f) ^ 0x7f)
		}
	case DATA_FMT_2S_COMPLEMENT:
		if raw&0x80 != 0 {
			raw = -((raw & 0x7f) ^ 0x7f) - 1
		}
	}
	raw1 := (float64(s.m)*float64(raw) + (float64(s.b) * math.Pow(10, float64(s.k1)))) * math.Pow(10, float64(s.k2))
	switch s.linearization & 0x7f {
	case L_LN:
		return math.Log(raw1), nil
	case L_LOG:
		return math.Log10(raw1), nil
	case L_LOG2:
		return math.Log2(raw1), nil
	case L_E:
		return math.Exp(raw1), nil
	case L_EXP10:
		return math.Pow(10, raw1), nil
	case L_EXP2:
		return math.Pow(2, raw1), nil
	case L_1_X:
		return 1.0 / raw1, nil
	case L_SQR:
		return math.Pow(raw1, 2), nil
	case L_CUBE:
		return math.Pow(raw1, 3), nil
	case L_SQRT:
		return math.Sqrt(raw1), nil
	case L_CUBERT:
		return math.Pow(raw1, 1.0/3), nil
	case L_LINEAR:
		return raw1, nil
	default:
		return 0, errors.Errorf("unknown linearization %d", s.linearization&0x7f)
	}
}
