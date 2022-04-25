import * as React from "react";

export default class Site extends React.Component {
  constructor(props) {
    super(props);
  }

  pause() {
    // noop
  }

  resume() {
    // noop
  }

  render() {
    return <iframe frameBorder="0" src={this.props.data.url} />;
  }
}
