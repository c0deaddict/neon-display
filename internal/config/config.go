package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type NatsConfig struct {
	Url          string `yaml:"url"`
	Username     string `yaml:"username",omitempty`
	PasswordFile string `yaml:"passwordFile",omitempty`
}

type MidiConfig struct {
	Input         string `yaml:"input"`
	Output        string `yaml:"output"`
	Channel       uint8  `yaml:"channel"`
	MaxInputValue uint   `yaml:"maxInputValue"`
}

type PulseAudioTargetType string

const (
	PlaybackStream PulseAudioTargetType = "PlaybackStream"
	RecordStream                        = "RecordStream"
	Sink                                = "Sink"
	Source                              = "Source"
)

type PulseAudioTarget struct {
	Type     PulseAudioTargetType `yaml:"type"`
	Name     string               `yaml:"name"`
	Mute     *uint8               `yaml:"mute",omitempty`
	Presence *uint8               `yaml:"presence",omitempty`
	Volume   *uint8               `yaml:"volume",omitempty`
}

type PulseAudioConfig struct {
	Targets []PulseAudioTarget `yaml:"targets"`
}

type Action struct {
	Type   string                 `yaml:"type"`
	Config map[string]interface{} `yaml:"config"`
}

type Config struct {
	Nats       NatsConfig       `yaml:"nats"`
	Midi       MidiConfig       `yaml:"midi"`
	PulseAudio PulseAudioConfig `yaml:"pulseaudio"`
	Actions    []Action         `yaml:"actions"`
}

func Read(filename string) (*Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
