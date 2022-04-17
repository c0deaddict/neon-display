import * as React from 'react';
import * as ReactDOM from 'react-dom';

// idea: compose one or more widgets on screen.
//
// react widget:
// - iframe (grafana dashboard, weather etc.)
// - photo slideshow
// - do something with the leds ?
// - ??

class App extends React.Component {
  constructor(props) {
    super(props);
  }

  componentDidMount() {
    const url = new URL('/ws', window.location.href);
    url.protocol = url.protocol.replace('http', 'ws');

    this.client = new WebSocket(url.href);
    this.client.addEventListener('open', (event) => {
      console.log('WebSocket client connected');
    });
    this.client.addEventListener('message', this.handleMessage.bind(this));

    this.timer = setInterval(() => {
      const id = self.crypto.randomUUID();
      const method = "ping";
      const params = null;
      this.client.send(JSON.stringify({id, method, params}));
    }, 5000);
  }

  componentWillUnmount() {
    this.client.close();
    clearInterval(this.timer);
  }

  handleMessage(message) {
    const message = JSON.parse(message.data);
    console.log(this, message);
  }

  render() {
    return (
      <div>
        <div id="message"></div>
        <div id="content">
          <img />
          <iframe frameBorder="0" src="" hidden></iframe>
        </div>      
      </div>
    );
  }
}

ReactDOM.render(
  <App />,
  document.getElementById('root')
);
