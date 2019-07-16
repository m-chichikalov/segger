## Atempt to Implement RTT and SystemView libs to use with TinyGo on the embedded targets.
July 14, 2019 - Done only simple "prove of concept".
The folder `examples/simple_rtt` contains simple example of usage.
stm32f4discovery was used to run code. 

To compile example use `tinygo build -o example.hex -gc=dumb -size=short -target=stm32f4discovery .`

The result should be:

![RTT with TinyGo!](https://user-images.githubusercontent.com/23377892/61259682-3c70fe00-a749-11e9-9563-e37c05c8861f.png)


- [TinyGo github REPO](https://github.com/tinygo-org/tinygo)
- [TinyGo](https://tinygo.org/)
