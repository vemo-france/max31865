package max31865

import (
	"math"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/host"
)

const (
	_CONFIG_REG    uint8 = 0x00
	_RTDMSB_REG    uint8 = 0x01
	_RTDLSB_REG    uint8 = 0x02
	_HFAULTMSB_REG uint8 = 0x03
	_HFAULTLSB_REG uint8 = 0x04
	_LFAULTMSB_REG uint8 = 0x05
	_LFAULTLSB_REG uint8 = 0x06
	_FAULTSTAT_REG uint8 = 0x07
)

const (
	_FAULT_HIGHTHRESH uint8 = 0x80
	_FAULT_LOWTHRESH  uint8 = 0x40
	_FAULT_REFINLOW   uint8 = 0x20
	_FAULT_REFINHIGH  uint8 = 0x10
	_FAULT_RTDINLOW   uint8 = 0x08
	_FAULT_OVUV       uint8 = 0x04
)

const (
	_CONFIG_BIAS      uint8 = 0x80
	_CONFIG_MODEAUTO  uint8 = 0x40
	_CONFIG_MODEOFF   uint8 = 0x00
	_CONFIG_1SHOT     uint8 = 0x20
	_CONFIG_3WIRE     uint8 = 0x10
	_CONFIG_24WIRE    uint8 = 0x00
	_CONFIG_FAULTSTAT uint8 = 0x02
	_CONFIG_FILT50HZ  uint8 = 0x01
	_CONFIG_FILT60HZ  uint8 = 0x00
)

const (
	_RTD_A float32 = 3.9083e-3
	_RTD_B float32 = -5.775e-7
)

const (
	WIRE_2 = 0
	WIRE_3 = 1
	WIRE_4 = 0
)

type Sensor struct {
	csPin   gpio.PinOut
	misoPin gpio.PinIn
	mosiPin gpio.PinOut
	clkPin  gpio.PinOut
}

func Init() error {
	if _, err := host.Init(); err != nil {
		return err
	}
	return nil
}

func Create(cs string, miso string, mosi string, clk string) *Sensor {

	s := &Sensor{}

	s.csPin = gpioreg.ByName(cs)
	s.misoPin = gpioreg.ByName(miso)
	s.mosiPin = gpioreg.ByName(mosi)
	s.clkPin = gpioreg.ByName(clk)

	s.csPin.Out(gpio.High)
	s.clkPin.Out(gpio.Low)

	s.setWires(WIRE_4)
	s.enableBias(false)
	s.autoConvert(false)
	s.clearFault()

	return s
}

func (s *Sensor) readFault() uint8 {
	return s.read8(_FAULTSTAT_REG)
}

func (s *Sensor) clearFault() {
	t := s.read8(_CONFIG_REG)
	t &= ^uint8(0x2C)
	t |= _CONFIG_FAULTSTAT
	s.write8(_CONFIG_REG, t)
}

func (s *Sensor) enableBias(b bool) {
	t := s.read8(_CONFIG_REG)
	if b {
		t |= _CONFIG_BIAS
	} else {
		t &= ^_CONFIG_BIAS
	}
	s.write8(_CONFIG_REG, t)
}

func (s *Sensor) autoConvert(b bool) {
	t := s.read8(_CONFIG_REG)
	if b {
		t |= _CONFIG_MODEAUTO
	} else {
		t &= ^_CONFIG_MODEAUTO
	}
	s.write8(_CONFIG_REG, t)
}

func (s *Sensor) setWires(wires int) {
	t := s.read8(_CONFIG_REG)
	if wires == WIRE_3 {
		t |= _CONFIG_3WIRE
	} else {
		t &= ^_CONFIG_3WIRE
	}
	s.write8(_CONFIG_REG, t)
}

func (s *Sensor) ReadTemperature(RTDnominal float32, refResistor float32) float32 {

	Rt := float32(s.ReadRTD())
	Rt /= 32768
	Rt *= refResistor

	Z1 := -_RTD_A
	Z2 := _RTD_A*_RTD_A - (4 * _RTD_B)
	Z3 := (4 * _RTD_B) / RTDnominal
	Z4 := 2 * _RTD_B

	temp := Z2 + (Z3 * Rt)
	temp = (float32(math.Sqrt(float64(temp))) + Z1) / Z4

	if temp >= 0 {
		return temp
	}

	Rt /= RTDnominal
	Rt *= 100

	rpoly := Rt

	temp = -242.02
	temp += 2.2228 * rpoly
	rpoly *= Rt
	temp += 2.5859e-3 * rpoly
	rpoly *= RTDnominal
	temp -= 4.8260e-6 * rpoly
	rpoly *= Rt
	temp -= 2.8183e-8 * rpoly
	rpoly *= Rt
	temp += 1.5243e-10 * rpoly

	return temp
}

func (s *Sensor) ReadRTD() uint16 {

	s.clearFault()
	s.enableBias(true)
	time.Sleep(10 * time.Millisecond)

	t := s.read8(_CONFIG_REG)
	t |= _CONFIG_1SHOT
	s.write8(_CONFIG_REG, t)
	time.Sleep(65 * time.Millisecond)

	rtd := s.read16(_RTDMSB_REG)
	rtd >>= 1
	return rtd

}

func (s *Sensor) write(addr uint8, v []uint8) {
	addr |= 0x80
	s.clkPin.Out(gpio.Low)
	s.csPin.Out(gpio.Low)
	s.transfer8(addr)
	for n := 0; n < len(v); n++ {
		s.transfer8(v[n])
	}
	s.csPin.Out(gpio.High)
}

func (s *Sensor) write8(addr uint8, v uint8) {
	s.write(addr, []uint8{v})
}

func (s *Sensor) read(addr uint8, v []uint8) {
	addr &= 0x7F
	s.clkPin.Out(gpio.Low)
	s.csPin.Out(gpio.Low)
	s.transfer8(addr)
	for n := 0; n < len(v); n++ {
		v[n] = s.transfer8(0xFF)
	}
	s.csPin.Out(gpio.High)
}

func (s *Sensor) read8(addr uint8) uint8 {
	var v = make([]uint8, 1)
	s.read(addr, v)
	return v[0]
}

func (s *Sensor) read16(addr uint8) uint16 {
	var v = make([]uint8, 2)
	s.read(addr, v)
	return (uint16(v[0]) << 8) | uint16(v[1])
}

func (s *Sensor) transfer8(v uint8) uint8 {
	var reply uint8 = 0
	for i := 7; i >= 0; i-- {
		reply <<= 1
		s.clkPin.Out(gpio.High)
		bv := v & (1 << uint(i))
		s.mosiPin.Out(bv != 0)
		s.clkPin.Out(gpio.Low)
		if s.misoPin.Read() {
			reply |= 1
		}
	}
	return reply
}
