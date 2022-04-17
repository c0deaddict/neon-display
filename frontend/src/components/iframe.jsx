import * as React from "react";

export default class Iframe extends React.Component {
  constructor(props) {
    super(props);
  }

  render() {
    return <iframe frameBorder="0" src={this.props.url} />;
  }
}
