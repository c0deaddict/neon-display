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


## Implementation

- Golang
- Split into two parts: HAL and Display. HAL watches the GPIO pins for the buttons and the PIR sensor. HAL also drives the LED strip. It needs root privileges for these operations. Display runs as a regular user, spawned by Cage (link). It communicates with HAL via gRPC over a unix socket.
- Cage starts a Firefox or Chromium kiosk browser and have it load the [frontend](./frontend). The frontend connects back to the Display process over Websocket to receive instructions on what to show.

## TODO

- allow photo albums and videos to have a Order in the filename, eg. "100_Hondjes" or "200_Video title.webm"
- reconnect to websocket if connection is lost
- read exif information photos (and cache this, test perf on pi)

## Debugging

```
journalctl -u neon-display-hal -f
journalctl -u neon-display -f
journalctl -t cage -f
```
