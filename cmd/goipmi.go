package main

import (
	"flag"
	"fmt"
	"github.com/neo-hu/goipmi"
	"github.com/olekukonko/tablewriter"
	"os"
)

func main() {
	var sdr bool
	var sel bool
	flag.BoolVar(&sdr, "sdr", sdr, "Print Sensor Data Repository entries and readings")
	flag.BoolVar(&sel, "sel", sel, "Print System Event Log")
	flag.Parse()
	t := goipmi.NewLocalIPMI()
	if err := t.Open(); err != nil {
		panic(err)
	}
	defer t.Close()

	table := tablewriter.NewWriter(os.Stdout)
	table.SetBorder(true)
	table.SetAutoWrapText(true)
	table.SetBorders(tablewriter.Border{Left: true, Top: true, Right: true, Bottom: true})
	table.SetCenterSeparator("|")
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	if sdr {
		table.SetHeader([]string{"SEL", "Value"})
		if err := t.SdrRepositoryEntries(func(name string, val *float64, unitCode uint8, unit string,
			sensorTypeCode, entityInstance uint8, sensorType string, err error) {
			table.Append([]string{
				name,
				fmt.Sprintf("%.2f %s", *val, unit),
			})
		}); err != nil {
			panic(err)
		}
	} else if sel {
		oem, err := t.GetOem()
		if err != nil {
			panic(err)
		}
		table.SetHeader([]string{"RecordId", "Timestamp", "Sensor", "Event", "Event Dir"})
		err = t.SelEntries(func(entry []byte) bool {
			e, err := goipmi.UnmarshalSelBinary(entry)
			if err != nil {
				return true
			}
			row := []string{
				fmt.Sprintf("%4x", e.RecordId),
				"",
				"",
				"",
				"",
			}

			if e.StandardType != nil {
				row[1] = e.StandardType.Timestamp.String()
				row[2] = e.StandardType.GenericSensorType()
				if e.StandardType.SensorNum > 0 {
					row[2] = fmt.Sprintf("%s #0x%02x", row[2], e.StandardType.SensorNum)
				}
				evt := e.StandardType.GetEventSensorType(oem)
				if evt != nil {
					row[3] = evt.Desc
				}
				row[4] = e.StandardType.GetEventDirString()
			} else if e.OemTsType != nil {
				row[1] = e.OemTsType.Timestamp.String()
				row[2] = fmt.Sprintf("OEM record %02x", e.RecordType)
				row[3] = fmt.Sprintf("%02x%02x%02x", e.OemTsType.ManfId[0], e.OemTsType.ManfId[1], e.OemTsType.ManfId[2])
				for _, b := range e.OemTsType.OemDefined {
					row[4] = fmt.Sprintf("%s%02x", row[4], b)
				}
			}
			table.Append(row)
			return true
		})
		if err != nil {
			panic(err)
		}
	}
	table.Render()
}
