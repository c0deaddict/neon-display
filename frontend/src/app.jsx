import * as React from "react";
import * as ReactDOM from "react-dom";

import Message from "./components/message.jsx";
import Photos from "./components/photos.jsx";
import Video from "./components/video.jsx";
import Site from "./components/site.jsx";

class App extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      content: null,
      message: null,
    };
    this.contentRef = React.createRef();
  }

  componentDidMount() {
    this.connect();
  }

  connect() {
    const url = new URL("/ws", window.location.href);
    url.protocol = url.protocol.replace("http", "ws");
    this.client = new WebSocket(url.href);
    this.client.addEventListener("open", (event) => {
      console.log("webSocket client connected");
    });
    this.client.addEventListener("message", this.handleMessage.bind(this));
    this.client.addEventListener("close", () => {
      if (this.client != null) {
        console.log("websocket connection lost, trying to re-connect");
        setTimeout(this.connect.bind(this), 1000);
      }
    });
    this.client.addEventListener("error", (err) => {
      console.error("websocket encountered error:", err.message);
      this.client.close();
    });
  }

  componentWillUnmount() {
    const ws = this.client;
    this.client = null; // Indicates we really want to close the socket.
    ws.close();
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
        this.showMessage({
          text: command.data.title,
          color: "red",
          show_seconds: 5,
        });
        this.showContent(command.data);
        break;

      case "pause_content":
        this.contentRef.current.pause();
        break;

      case "resume_content":
        this.contentRef.current.resume();
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

  showContent(data) {
    this.setState({ content: data });
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
    }

    switch (this.state.content.type) {
      case "photos":
        return <Photos data={this.state.content.data} ref={this.contentRef} />;
      case "video":
        return <Video data={this.state.content.data} ref={this.contentRef} />;
      case "site":
        return <Site data={this.state.content.data} ref={this.contentRef} />;
      default:
        console.error("unknown content type: " + this.state.content.type);
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
