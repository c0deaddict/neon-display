import * as React from "react";
import ProgressBar from "./progress_bar.jsx";

export default class Photos extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      index: 0,
    };
  }

  getBackgroundColor() {
    return this.props.data.background_color || "black";
  }

  getProgress() {
    return (this.state.index + 1) / this.props.data.photos.length;
  }

  startTimer() {
    this.timer = setInterval(
      this.nextPhoto.bind(this),
      this.props.data.delay_seconds * 1000
    );
  }

  stopTimer() {
    clearInterval(this.timer);
  }

  componentDidMount() {
    this.startTimer();
  }

  componentWillUnmount() {
    this.stopTimer();
  }

  nextPhoto() {
    const index = (this.state.index + 1) % this.props.data.photos.length;
    this.setState({ index });
  }

  pause() {
    this.stopTimer();
  }

  resume() {
    this.startTimer();
  }

  render() {
    const photo = this.props.data.photos[this.state.index];
    return (
      <div
        className="photos"
        style={{ backgroundColor: this.getBackgroundColor() }}
      >
        <img src={"/photo/" + photo.image_path} />
        <div className="photo-info">
          <div className="photo-datetime">{photo.datetime}</div>
          <div className="photo-description">{photo.description}</div>
          <div className="photo-camera">{photo.camera}</div>
        </div>
        <ProgressBar progress={this.getProgress()} />
      </div>
    );
  }
}
