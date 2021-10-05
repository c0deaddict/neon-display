# Neon Display

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

exec feh --hide-pointer -x -q -D 5 -B black -Z -z -F --auto-rotate -R 5 -r /pics
```

Switching images can be done by sending a signal to feh:
```
SIGUSR1  Switch to next image
SIGUSR2  Switch to previous image
```

## Improvements

- If chromium loses the web socket connection (no more pongs) then reload the page.
- Add an admin interface
- Auto rotate photos

## Setup

```sh
cd controller
pip3 install -r requirements.txt
```

# Autologin

```txt
sudo raspi-config
boot options
desktop/cli
console, automatically login as user pi
```

This will run `/home/pi/.profile` after boot, which will `startx`, which will
execute our `.xinitrc`.
