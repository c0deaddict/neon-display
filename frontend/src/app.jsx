import * as React from "react";
import * as ReactDOM from "react-dom";
import Message from "./components/message.jsx";
import Slideshow from "./components/slideshow.jsx";
import Iframe from "./components/iframe.jsx";

class App extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      content: null,
      message: null,
    };
  }

  componentDidMount() {
    const url = new URL("/ws", window.location.href);
    url.protocol = url.protocol.replace("http", "ws");

    this.client = new WebSocket(url.href);
    this.client.addEventListener("open", (event) => {
      console.log("WebSocket client connected");
    });
    this.client.addEventListener("message", this.handleMessage.bind(this));

    this.timer = setInterval(() => {
      const id = self.crypto.randomUUID();
      const method = "ping";
      const params = null;
      this.client.send(JSON.stringify({ id, method, params }));
    }, 5000);
  }

  componentWillUnmount() {
    this.client.close();
    clearInterval(this.timer);
  }

  handleMessage(wsMessage) {
    const message = JSON.parse(wsMessage.data);
    switch (message.type) {
      case "command":
        this.handleCommand(message.data);
        break;

      case "response":
        this.handleResponse(message.data);
        break;

      default:
        console.error("unexpected message type received:", message.type);
    }
  }

  handleCommand(command) {
    switch (command.type) {
      case "show_content":
        this.startSlideshow(command.data);
        break;

      case "pause_content_":
        this.stopSlideshow();
        break;

      case "resume_content":
        this.openUrl(command.data);
        break;

      case "show_message":
        this.showMessage(command.data);
        break;

      case "reload":
        window.location.reload();
        break;

      default:
        console.error("unexpected command received:", command.type);
    }
  }

  startSlideshow(data) {
    this.setState({ content: { type: "slideshow", data } });
  }

  stopSlideshow() {
    this.setState({ content: null });
  }

  openUrl(data) {
    this.setState({ content: { type: "iframe", data } });
  }

  showMessage(data) {
    this.setState({ message: data });
    setTimeout(
      () => this.setState({ message: null }),
      data.show_seconds * 1000
    );
  }

  handleResponse(response) {
    console.log("handleResponse is not implemented", response);
  }

  renderContent() {
    if (this.state.content == null) {
      return null;
    } else if (this.state.content.type == "slideshow") {
      return <Slideshow slideshow={this.state.content.data} />;
    } else if (this.state.content.type == "iframe") {
      return <Iframe url={this.state.content.data.url} />;
    } else {
      return null;
    }
  }

  render() {
    return (
      <div className="app">
        {this.state.message ? <Message message={this.state.message} /> : null}
        {this.renderContent()}
      </div>
    );
  }
}

ReactDOM.render(<App />, document.getElementById("root"));
