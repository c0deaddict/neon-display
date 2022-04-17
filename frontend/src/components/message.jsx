import * as React from "react";

export default class Message extends React.Component {
  constructor(props) {
    super(props);
  }

  render() {
    return (
      <div className="message" style={{ color: this.props.message.color }}>
        {this.props.message.text}
      </div>
    );
  }
}
