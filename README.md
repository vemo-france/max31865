vemo-france/max31865
====================

This is a golang/periph port of [Adafruit's MAX31865 library](https://github.com/adafruit/Adafruit_MAX31865). It's used to drive a [MAX31865 ic](https://datasheets.maximintegrated.com/en/ds/MAX31865.pdf) or board such as [Adafruit's MAX31865 breakout board](https://www.adafruit.com/product/3328) on any device supported by [periph](https://github.com/google/periph). That ic is a PT100/PT1000 (precision temperature sensor) amplifier.


Installation
------------

````
go get github.com/vemo-france/max31865
````

Usage
----

````Go
package main

import (
	"fmt"
	"log"

	"github.com/vemo-france/max31865"
)

func main() {

	if err := max31865.Init(); err != nil {
		log.Fatalf("initialization failed : %s", err)
	}

	// pass in pin names as defined in periph
	sensor := max31865.Create("8", "9", "10", "11")

	// 100 is sensor's resistance at 0°C in ohms (PT100 -> 100, PT1000 -> 1000)
	// 430 is reference resistance in ohms (430 for adafruit's board)
	temp := sensor.ReadTemperature(100, 430)

	fmt.Printf("Temperature is %f°C\n", temp)
}
````

TODO
----

- Handle 2, 3 et 4 wires sensors
- Create and ReadTemperature can fail, so we should return `( ??, error)` to handle that