# Neon Display

-- Use `firefox -kiosk URL` instead of Chromium.

## Hardware
Neon is a Raspberry Pi attached to a monitor for displaying photos, graphs, etc.

It has attached:
- PIR motion sensor on GPIO 17
- Red push button (pulled down) on GPIO 22
- Yellow push button (pulled down) on GPIO 27
- WS2812 LED strip at GPIO 12

## Functional requirements
1. When there is no motion for 2 minutes, turn off the display and stop the slideshow.
2. When motion is detected, turn on the display and start the slideshow.
3. When pressing the red button, goto the next photo.
4. When pressing the yellow button, goto the previous photo.

### Extra
5. Add controls/configure settings
    - perhaps via HTTP? or MQTT?
    - configure slideshow interval
    - configure no motion timeout
    - configure button behaviour?
6. Publish motion sensor to MQTT.
7. Add more narrowcasting sweetness:
    - render web pages (power monitoring)
    - youtube videos?
    - and more!

## Display

Turning it off:
```sh
tvservice -o
```

Turning it on:
```sh
export DISPLAY=:0  # needed by xset
tvservice --preferred && xset -dpms off
```

## Slideshow

~/.xinitrc:

```sh
#!/bin/sh
xset -dpms
xset s off
```

## Improvements

- If chromium loses the web socket connection (no more pongs) then reload the page.
- Add an admin interface
- Auto rotate photos

## Setup

### LEDS

https://github.com/jgarff/rpi_ws281x

/etc/modprobe.d/snd-blacklist.conf:

```
blacklist snd_bcm2835
```

Clone the git repo and install `scons`, then:

```
scons
cp *.a /usr/local/lib/
cp *.h /usr/local/include/
```

Now the go app can be build on the Pi.
