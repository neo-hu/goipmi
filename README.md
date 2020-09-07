# goipmi
golang ipmi采集

<pre>
t := goipmi.NewLocalIPMI()
if err := t.Open(); err != nil {
    panic(err)
}
defer t.Close()

## sdr 设备传感器采集
err := t.SdrRepositoryEntries(func(name string, val *float64, unitCode uint8, unit string,
			sensorTypeCode, entityInstance uint8, sensorType string, err error) {
    table.Append([]string{
        name,
        fmt.Sprintf("%.2f %s", *val, unit),
    })
});
    
## sel 设备日志采集
err = t.SelEntries(func(entry []byte) bool {
    e, err := goipmi.UnmarshalSelBinary(entry)
    if err != nil {
        return true
    }
    return true
})

</pre>