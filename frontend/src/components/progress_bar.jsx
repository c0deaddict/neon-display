import * as React from "react";

export default class ProgressBar extends React.Component {
  constructor(props) {
    super(props);
  }

  render() {
    return (
      <div className="progress-bar">
        <div
          className="progress"
          style={{ width: this.props.progress * 100 + "%" }}
        ></div>
      </div>
    );
  }
}
