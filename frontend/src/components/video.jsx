import * as React from "react";

export default class Video extends React.Component {
  constructor(props) {
    super(props);
    this.videoRef = React.createRef();
  }

  pause() {
    this.videoRef.current.pause();
  }

  resume() {
    this.videoRef.current.play();
  }

  render() {
    return (
      <video ref={this.videoRef} autoPlay={true} loop={true}>
        <source src={"/video/" + this.props.data.path} />
      </video>
    );
  }
}
